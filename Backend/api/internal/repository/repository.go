package repository

// UserRepository kullanıcı verisine erişim katmanı.
type UserRepository struct{}

func NewUserRepository() *UserRepository { return &UserRepository{} }

// LinkRepository link verisine erişim katmanı.
type LinkRepository struct{}

func NewLinkRepository() *LinkRepository { return &LinkRepository{} }
