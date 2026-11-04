package bearer

import (
	"context"
	"net/http"
	"strings"

	"github.com/adk-saugat/rally/gateway/internal/utils/grpcmeta"
	"github.com/gin-gonic/gin"
)

func Context(ginContext *gin.Context) (context.Context, bool) {
	header := strings.TrimSpace(ginContext.GetHeader("Authorization"))
	token, hasBearer := strings.CutPrefix(header, "Bearer ")
	token = strings.TrimSpace(token)
	if !hasBearer || token == "" {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}

	return grpcmeta.ContextWithAuthorization(ginContext.Request.Context(), token), true
}
