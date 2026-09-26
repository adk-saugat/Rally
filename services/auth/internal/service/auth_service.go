package service

import (
	"context"
	"errors"
	"strings"

	"github.com/adk-saugat/rally/services/auth/internal/repository"
	"github.com/adk-saugat/rally/services/auth/internal/utils"
	"github.com/adk-saugat/rally/shared/types"
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
		return "", types.ErrInvalidInput
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	userID, err := auth.users.CreateUser(ctx, email, passwordHash, firstName)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (auth *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "", types.ErrInvalidInput
	}

	user, err := auth.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return "", types.ErrInvalidCredentials
		}
		return "", err
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return "", types.ErrInvalidCredentials
	}

	return utils.SignAccessToken(auth.jwtSecret, user.ID)
}

func (auth *AuthService) GetUsers(ctx context.Context, userIDs []string) ([]*types.User, error) {
	ids := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			ids = append(ids, userID)
		}
	}

	return auth.users.GetUsersByIDs(ctx, ids)
}
