package handler

import (
	"context"

	authv1 "github.com/adk-saugat/rally/gen/auth/v1"
	"github.com/adk-saugat/rally/services/auth/internal/service"
	"github.com/adk-saugat/rally/shared/types"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (handler *AuthHandler) SignUp(ctx context.Context, request *authv1.SignUpRequest) (*authv1.SignUpResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	userID, err := handler.auth.SignUp(ctx, request.GetFirstName(), request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, types.GRPCError(err)
	}

	return &authv1.SignUpResponse{UserId: userID}, nil
}

func (handler *AuthHandler) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	token, err := handler.auth.Login(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, types.GRPCError(err)
	}

	return &authv1.LoginResponse{Token: token}, nil
}

func (handler *AuthHandler) GetUsers(ctx context.Context, request *authv1.GetUsersRequest) (*authv1.GetUsersResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	users, err := handler.auth.GetUsers(ctx, request.GetUserIds())
	if err != nil {
		return nil, types.GRPCError(err)
	}

	responseUsers := make([]*authv1.User, 0, len(users))
	for _, user := range users {
		responseUsers = append(responseUsers, &authv1.User{
			UserId:    user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
		})
	}

	return &authv1.GetUsersResponse{Users: responseUsers}, nil
}
