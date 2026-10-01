package clients

import (
	"context"

	groupv1 "github.com/adk-saugat/rally/gen/group/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GroupClient struct {
	connection *grpc.ClientConn
	groups     groupv1.GroupServiceClient
}

func NewGroupClient(address string) (*GroupClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GroupClient{
		connection: connection,
		groups:     groupv1.NewGroupServiceClient(connection),
	}, nil
}

func (client *GroupClient) IsMember(ctx context.Context, groupID string) (bool, error) {
	response, err := client.groups.IsMember(outgoingWithAuthorization(ctx), &groupv1.IsMemberRequest{
		GroupId: groupID,
	})
	if err != nil {
		return false, err
	}
	return response.GetIsMember(), nil
}

func (client *GroupClient) Close() error {
	return client.connection.Close()
}

func outgoingWithAuthorization(ctx context.Context) context.Context {
	incomingMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	values := incomingMetadata.Get("authorization")
	if len(values) == 0 {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", values[0])
}
