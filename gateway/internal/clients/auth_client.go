package clients

import (
	"context"

	authv1 "github.com/adk-saugat/rally/gen/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	connection *grpc.ClientConn
	auth       authv1.AuthServiceClient
}

func NewAuthClient(address string) (*AuthClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		connection: connection,
		auth:       authv1.NewAuthServiceClient(connection),
	}, nil
}

func (client *AuthClient) SignUp(ctx context.Context, firstName, email, password string) (string, error) {
	response, err := client.auth.SignUp(ctx, &authv1.SignUpRequest{
		FirstName: firstName,
		Email:     email,
		Password:  password,
	})
	if err != nil {
		return "", err
	}
	return response.GetUserId(), nil
}

func (client *AuthClient) Login(ctx context.Context, email, password string) (string, error) {
	response, err := client.auth.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return response.GetToken(), nil
}

func (client *AuthClient) Close() error {
	return client.connection.Close()
}
