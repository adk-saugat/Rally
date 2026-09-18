package handler

import (
	"context"
	"errors"

	authv1 "github.com/adk-saugat/rally/gen/auth/v1"
	"github.com/adk-saugat/rally/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userID, err := handler.auth.SignUp(ctx, request.GetFirstName(), request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.SignUpResponse{UserId: userID}, nil
}

func (handler *AuthHandler) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	token, err := handler.auth.Login(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LoginResponse{Token: token}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
