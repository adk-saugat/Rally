package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeError(context *gin.Context, err error) {
	if err == nil {
		return
	}

	grpcStatus, ok := status.FromError(err)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	context.JSON(httpStatus(grpcStatus.Code()), gin.H{"error": grpcStatus.Message()})
}

func httpStatus(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
