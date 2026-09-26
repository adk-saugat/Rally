package clients

import (
	"context"

	authv1 "github.com/adk-saugat/rally/gen/auth/v1"
	"github.com/adk-saugat/rally/shared/types"
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

func (client *AuthClient) GetUsers(ctx context.Context, userIDs []string) ([]*types.User, error) {
	response, err := client.auth.GetUsers(ctx, &authv1.GetUsersRequest{UserIds: userIDs})
	if err != nil {
		return nil, err
	}

	users := make([]*types.User, 0, len(response.GetUsers()))
	for _, user := range response.GetUsers() {
		users = append(users, &types.User{
			ID:        user.GetUserId(),
			Email:     user.GetEmail(),
			FirstName: user.GetFirstName(),
		})
	}
	return users, nil
}

func (client *AuthClient) Close() error {
	return client.connection.Close()
}
