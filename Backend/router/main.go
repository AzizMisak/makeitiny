package main

import (
	"log"
	"net/http"
	"os"

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

	linkRepo := repository.NewLinkRepository()
	h := handler.New(service.NewRedirectService(linkRepo))

	log.Println("router listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, gen.Handler(h)))
}
