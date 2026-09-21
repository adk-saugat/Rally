package utils

import (
	"context"
	"strings"

	"github.com/adk-saugat/rally/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey struct{}

var userIDKey = contextKey{}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}

func UnaryAuth(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, err := userIDFromMetadata(ctx, jwtSecret)
		if err != nil {
			return nil, types.GRPCError(types.ErrUnauthorized)
		}
		return handler(ContextWithUserID(ctx, userID), request)
	}
}

func userIDFromMetadata(ctx context.Context, jwtSecret string) (string, error) {
	incomingMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", types.ErrUnauthorized
	}

	values := incomingMetadata.Get("authorization")
	if len(values) == 0 {
		return "", types.ErrUnauthorized
	}

	header := strings.TrimSpace(values[0])
	tokenString, hasBearer := strings.CutPrefix(header, "Bearer ")
	if !hasBearer {
		return "", types.ErrUnauthorized
	}

	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", types.ErrUnauthorized
	}

	return ParseAccessToken(jwtSecret, tokenString)
}
