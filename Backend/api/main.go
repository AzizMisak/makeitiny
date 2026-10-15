package main

import (
	"log"
	"net/http"
	"os"

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

	userRepo := repository.NewUserRepository()
	linkRepo := repository.NewLinkRepository()
	h := handler.New(service.NewAuthService(userRepo), service.NewLinkService(linkRepo))

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
