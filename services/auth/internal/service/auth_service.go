package service

import (
	"context"
	"errors"
	"strings"

	"github.com/adk-saugat/rally/services/auth/internal/repository"
	"github.com/adk-saugat/rally/services/auth/internal/utils"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailTaken         = repository.ErrEmailTaken
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (auth *AuthService) SignUp(ctx context.Context, firstName, email, password string) (string, error) {
	firstName = strings.TrimSpace(firstName)
	email = strings.TrimSpace(email)
	if firstName == "" || email == "" || password == "" {
		return "", ErrInvalidInput
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	userID, err := auth.users.CreateUser(ctx, email, passwordHash, firstName)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return "", ErrEmailTaken
		}
		return "", err
	}
	return userID, nil
}

func (auth *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "", ErrInvalidInput
	}

	user, err := auth.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	return utils.SignAccessToken(auth.jwtSecret, user.ID)
}
