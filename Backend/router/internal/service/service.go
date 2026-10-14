package service

import "makeitiny/router/internal/repository"

// RedirectService yönlendirme, kilit açma ve önizleme iş mantığı.
type RedirectService struct {
	links *repository.LinkRepository
}

func NewRedirectService(links *repository.LinkRepository) *RedirectService {
	return &RedirectService{links: links}
}
