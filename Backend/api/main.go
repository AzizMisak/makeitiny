package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	"makeitiny/api/internal/db"
	"makeitiny/api/internal/gen"
	"makeitiny/api/internal/handler"
	"makeitiny/api/internal/repository"
	"makeitiny/api/internal/service"
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

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
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

	queries := db.New(pool)
	userRepo := repository.NewUserRepository(queries, rdb)
	linkRepo := repository.NewLinkRepository(queries, rdb)

	authSvc := service.NewAuthService(userRepo, jwtSecret)
	linkSvc := service.NewLinkService(linkRepo)
	h := handler.New(authSvc, linkSvc)

	allowedOrigins := []string{
		"http://localhost:8082",
		"http://127.0.0.1:8082",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost",
		"http://127.0.0.1",
	}

	handlerWithCORS := withCORS(gen.Handler(h), allowedOrigins)

	log.Println("api listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
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
