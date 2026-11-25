package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/handler"
	"makeitiny/api/internal/qr"
	"makeitiny/api/internal/repository"
	"makeitiny/api/internal/service"
	"makeitiny/api/internal/worker"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://makeitiny:makeitiny@localhost:5432/makeitiny?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "makeitiny-dev-secret-key-2026"
	}

	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost"
	}

	qrDir := os.Getenv("QR_STORAGE_DIR")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	poolCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("geçersiz DATABASE_URL: %v", err)
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("postgres bağlantı havuzu oluşturulamadı: %v", err)
	}
	defer pool.Close()

	for i := 1; i <= 10; i++ {
		if err = pool.Ping(ctx); err == nil {
			break
		}
		log.Printf("postgres bekleniyor (%d/10): %v", i, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("postgres'e ulaşılamadı: %v", err)
	}

	runMigrations(pool)

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("geçersiz REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("uyarı: redis ping başarısız (%v), servis çalışmaya devam ediyor", err)
	}

	qrStore, err := qr.NewStore(qrDir)
	if err != nil {
		log.Fatalf("qr depolama başlatılamadı: %v", err)
	}

	queries := db.New(pool)

	syncWorker := worker.NewSyncWorker(pool, queries, rdb, qrStore)
	syncWorker.Start(ctx)

	userRepo := repository.NewUserRepository(queries, rdb, qrStore)
	linkRepo := repository.NewLinkRepository(queries, rdb, qrStore)

	authSvc := service.NewAuthService(userRepo, jwtSecret)
	linkSvc := service.NewLinkService(linkRepo, qrStore, publicBaseURL)
	h := handler.New(authSvc, linkSvc)

	allowedOrigins := []string{
		"http://localhost:8082",
		"http://127.0.0.1:8082",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost",
		"http://127.0.0.1",
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(gen.HandlerFromMux(h, newPriorityMux()), allowedOrigins),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("api listening on :" + port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("api sunucu hatası: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("api kapanıyor, bekleyen işlemler boşaltılıyor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func runMigrations(pool *pgxpool.Pool) {
	candidates := []string{"/migrations", "../db/migrations", "./Backend/db/migrations"}
	var migDir string
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			migDir = c
			break
		}
	}
	if migDir == "" {
		log.Println("migrations dizini bulunamadı, otomatik goose migration atlandı")
		return
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Printf("goose dialect hatası: %v", err)
		return
	}
	if err := goose.Up(sqlDB, migDir); err != nil {
		log.Printf("goose up hatası: %v", err)
	}
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// priorityMux, Go 1.22+ ServeMux'ta /links/{id}/qr ile /links/check-alias/{alias} ve /links/logs/{id}
// arasındaki wildcard çakışmasını önlemek için QR rotalarını ayrı bir mux üzerinde tutar.
type priorityMux struct {
	mainMux *http.ServeMux
	qrMux   *http.ServeMux
}

func newPriorityMux() *priorityMux {
	return &priorityMux{
		mainMux: http.NewServeMux(),
		qrMux:   http.NewServeMux(),
	}
}

func (p *priorityMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if strings.HasSuffix(pattern, "/qr") {
		p.qrMux.HandleFunc(pattern, handler)
		return
	}
	p.mainMux.HandleFunc(pattern, handler)
}

func (p *priorityMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segs) == 3 && segs[0] == "links" && segs[2] == "qr" && segs[1] != "check-alias" && segs[1] != "logs" {
		p.qrMux.ServeHTTP(w, r)
		return
	}
	p.mainMux.ServeHTTP(w, r)
}
