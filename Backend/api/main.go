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

	log.Println("api listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, gen.Handler(h)))
}
