package clients

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func ContextWithAuthorization(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
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
