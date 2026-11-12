package bearer

import (
	"context"
	"net/http"
	"strings"

	"github.com/adk-saugat/rally/gateway/internal/utils/grpcmeta"
	"github.com/adk-saugat/rally/shared/utils"
	"github.com/gin-gonic/gin"
)

func Token(ginContext *gin.Context) (string, bool) {
	header := strings.TrimSpace(ginContext.GetHeader("Authorization"))
	token, hasBearer := strings.CutPrefix(header, "Bearer ")
	token = strings.TrimSpace(token)
	if !hasBearer || token == "" {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}

	return token, true
}

func Context(ginContext *gin.Context) (context.Context, bool) {
	token, ok := Token(ginContext)
	if !ok {
		return nil, false
	}
	return grpcmeta.ContextWithAuthorization(ginContext.Request.Context(), token), true
}

func AuthenticatedContext(ginContext *gin.Context, jwtSecret string) (context.Context, bool) {
	token, ok := Token(ginContext)
	if !ok {
		return nil, false
	}
	userID, err := utils.ParseAccessToken(jwtSecret, token)
	if err != nil {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}
	ctx := grpcmeta.ContextWithAuthorization(ginContext.Request.Context(), token)
	return utils.ContextWithUserID(ctx, userID), true
}
