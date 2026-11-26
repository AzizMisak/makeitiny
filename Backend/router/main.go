package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"makeitiny/router/internal/db"
	"makeitiny/router/internal/gen"
	"makeitiny/router/internal/handler"
	"makeitiny/router/internal/repository"
	"makeitiny/router/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://makeitiny:makeitiny@localhost:5432/makeitiny?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost"
	}

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
	linkRepo := repository.NewLinkRepository(queries, rdb)
	h := handler.New(service.NewRedirectService(linkRepo), frontendURL)

	allowedOrigins := []string{
		"http://localhost:8082",
		"http://127.0.0.1:8082",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost",
		"http://127.0.0.1",
	}
	if u, err := url.Parse(strings.TrimSpace(frontendURL)); err == nil && u.Scheme != "" && u.Host != "" {
		origin := u.Scheme + "://" + u.Host
		found := false
		for _, o := range allowedOrigins {
			if o == origin {
				found = true
				break
			}
		}
		if !found {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	routes := gen.HandlerWithOptions(h, gen.StdHTTPServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(gen.ApiError{
				Code:    "invalid_request",
				Message: err.Error(),
			})
		},
	})
	root := newRootHandler(h, routes)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(root, allowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("router listening on :" + port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("router sunucu hatası: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("router kapanıyor, açık istekler tamamlanıyor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("router kapanış uyarısı: %v", err)
	}
}

func newRootHandler(h *handler.Handler, routes http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kısa kod içermeyen istek (nginx üzerinden GET /r/) düz metin 404 yerine hata ekranına gider.
		if r.URL.Path == "/" || r.URL.Path == "" {
			h.RootNotFound(w, r)
			return
		}
		if (r.Method == http.MethodGet || r.Method == http.MethodHead) && strings.HasSuffix(r.URL.Path, "/") {
			trimmed := strings.TrimRight(r.URL.Path, "/")
			if trimmed == "" {
				h.RootNotFound(w, r)
				return
			}
			if !strings.Contains(strings.TrimPrefix(trimmed, "/"), "/") && trimmed != "/health" {
				r.URL.Path = trimmed
				if r.URL.RawPath != "" {
					r.URL.RawPath = strings.TrimRight(r.URL.RawPath, "/")
				}
			}
		}
		if r.Method == http.MethodGet &&
			strings.Count(strings.Trim(r.URL.Path, "/"), "/") >= 1 &&
			!strings.HasSuffix(r.URL.Path, "/preview") &&
			!strings.HasSuffix(r.URL.Path, "/wait") {
			h.RootNotFound(w, r)
			return
		}
		routes.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Add("Vary", "Origin")
		}
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
