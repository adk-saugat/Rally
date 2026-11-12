package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/adk-saugat/rally/gateway/internal/clients"
	"github.com/adk-saugat/rally/gateway/internal/utils/bearer"
	"github.com/adk-saugat/rally/gateway/internal/utils/httperr"
	"github.com/adk-saugat/rally/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type LocationHandler struct {
	jwtSecret string
	locations *clients.LocationClient
	upgrader  websocket.Upgrader
}

func NewLocationHandler(locations *clients.LocationClient, jwtSecret string) *LocationHandler {
	return &LocationHandler{
		jwtSecret: jwtSecret,
		locations: locations,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: 5 * time.Second,
		},
	}
}

// LocationStream is the entry point for a group's live location WebSocket.
// The gRPC relay is added next.
func (handler *LocationHandler) LocationStream(ginContext *gin.Context) {
	ctx, ok := bearer.Context(ginContext)
	if !ok {
		return
	}

	token := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ginContext.GetHeader("Authorization")), "Bearer "))
	userID, err := utils.ParseAccessToken(handler.jwtSecret, token)
	if err != nil {
		ginContext.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	groupID := strings.TrimSpace(ginContext.Param("group_id"))
	if groupID == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	ctx, cancel := context.WithCancel(utils.ContextWithUserID(ctx, userID))
	defer cancel()
	ginContext.Request = ginContext.Request.WithContext(ctx)
	ginContext.Set("group_id", groupID)

	stream, err := handler.locations.StreamLocation(ctx)
	if err != nil {
		httperr.Write(ginContext, err)
		return
	}
	defer stream.CloseSend()

	connection, err := handler.upgrader.Upgrade(ginContext.Writer, ginContext.Request, nil)
	if err != nil {
		return
	}
	defer connection.Close()

	// Until the relay is implemented, explicitly close after the handshake.
	_ = connection.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "location relay is not implemented yet"),
		time.Now().Add(time.Second))
}
