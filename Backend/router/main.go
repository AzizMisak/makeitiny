package main

import (
	"context"
	"log"
	"net/http"
	"os"
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

	handlerWithCORS := withCORS(gen.Handler(h), allowedOrigins)

	log.Println("router listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
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
