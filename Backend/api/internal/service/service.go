package service

import "makeitiny/api/internal/repository"

// AuthService kimlik doğrulama iş mantığı.
type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

// LinkService link iş mantığı.
type LinkService struct {
	links *repository.LinkRepository
}

func NewLinkService(links *repository.LinkRepository) *LinkService {
	return &LinkService{links: links}
}
