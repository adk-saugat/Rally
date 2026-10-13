package clients

import (
	"context"

	locationv1 "github.com/adk-saugat/rally/gen/location/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LocationClient struct {
	connection *grpc.ClientConn
	locations  locationv1.LocationServiceClient
}

func NewLocationClient(address string) (*LocationClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &LocationClient{
		connection: connection,
		locations:  locationv1.NewLocationServiceClient(connection),
	}, nil
}

func (client *LocationClient) StreamLocation(ctx context.Context) (locationv1.LocationService_StreamLocationClient, error) {
	return client.locations.StreamLocation(outgoingWithAuthorization(ctx))
}

func (client *LocationClient) Close() error {
	return client.connection.Close()
}
