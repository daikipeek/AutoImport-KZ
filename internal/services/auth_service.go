package services

import (
	"context"
	"errors"
	"regexp"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailExists = errors.New("email already registered")
var ErrInvalidInput = errors.New("invalid input")

// AuthService handles user registration and login.
type AuthService struct {
	users *repositories.UserRepo
}

func NewAuthService(users *repositories.UserRepo) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (models.User, error) {
	if name == "" || email == "" || password == "" || !emailRegex.MatchString(email) {
		return models.User{}, ErrInvalidInput
	}

	if _, err := s.users.ByEmail(ctx, email); err == nil {
		return models.User{}, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	}

	id, err := s.users.Create(ctx, user)
	if err != nil {
		return models.User{}, err
	}
	user.ID = id
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, error) {
	if email == "" || password == "" {
		return models.User{}, ErrInvalidInput
	}

	user, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return models.User{}, ErrInvalidCredentials
	}

	return user, nil
}
