package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adk-saugat/rally/gateway/internal/clients"
	"github.com/adk-saugat/rally/gateway/internal/utils/bearer"
	"github.com/adk-saugat/rally/gateway/internal/utils/httperr"
	locationv1 "github.com/adk-saugat/rally/gen/location/v1"
	"github.com/adk-saugat/rally/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Location messages and validation.
var errInvalidCoordinates = errors.New("invalid coordinates")

type locationCoordinates struct {
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	TimestampUnixMs int64    `json:"timestamp_unix_ms"`
}

type memberLocation struct {
	GroupID         string  `json:"group_id"`
	UserID          string  `json:"user_id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TimestampUnixMs int64   `json:"timestamp_unix_ms"`
}

func (coordinates locationCoordinates) locationUpdate(groupID, userID string) (*locationv1.LocationUpdate, error) {
	if coordinates.Latitude == nil || coordinates.Longitude == nil ||
		*coordinates.Latitude < -90 || *coordinates.Latitude > 90 ||
		*coordinates.Longitude < -180 || *coordinates.Longitude > 180 {
		return nil, errInvalidCoordinates
	}
	if coordinates.TimestampUnixMs == 0 {
		coordinates.TimestampUnixMs = time.Now().UnixMilli()
	}
	return &locationv1.LocationUpdate{
		GroupId: groupID, UserId: userID,
		Latitude: *coordinates.Latitude, Longitude: *coordinates.Longitude,
		TimestampUnixMs: coordinates.TimestampUnixMs,
	}, nil
}

// Handler setup and connection handling.
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
func (handler *LocationHandler) LocationStream(ginContext *gin.Context) {
	ctx, ok := bearer.AuthenticatedContext(ginContext, handler.jwtSecret)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(ginContext.Param("group_id"))
	if groupID == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := handler.locations.StreamLocation(ctx)
	if err != nil {
		httperr.Write(ginContext, err)
		return
	}

	connection, err := handler.upgrader.Upgrade(ginContext.Writer, ginContext.Request, nil)
	if err != nil {
		return
	}
	defer connection.Close()

	relayLocations(ctx, connection, stream, groupID, utils.UserIDFromContext(ctx), cancel)
}

// Bidirectional message relay.
func relayLocations(ctx context.Context, connection *websocket.Conn, stream locationv1.LocationService_StreamLocationClient, groupID, userID string, cancel context.CancelFunc) {
	connection.SetReadLimit(4096)
	done := make(chan error, 2)
	go func() {
		done <- receiveCoordinates(connection, stream, groupID, userID)
	}()
	go func() {
		done <- sendMemberLocations(connection, stream)
	}()

	completed := 0
	var err error
	select {
	case err = <-done:
		completed++
	case <-ctx.Done():
		err = ctx.Err()
	}
	closeLocationConnection(connection, err)
	cancel()
	_ = connection.Close()
	for ; completed < 2; completed++ {
		<-done
	}
}

func receiveCoordinates(connection *websocket.Conn, stream locationv1.LocationService_StreamLocationClient, groupID, userID string) error {
	defer stream.CloseSend()
	for {
		var coordinates locationCoordinates
		if err := connection.ReadJSON(&coordinates); err != nil {
			var syntaxError *json.SyntaxError
			var typeError *json.UnmarshalTypeError
			if errors.As(err, &syntaxError) || errors.As(err, &typeError) || errors.Is(err, io.ErrUnexpectedEOF) {
				return errInvalidCoordinates
			}
			return err
		}
		update, err := coordinates.locationUpdate(groupID, userID)
		if err != nil {
			return err
		}
		if err := stream.Send(update); err != nil {
			return err
		}
	}
}

func sendMemberLocations(connection *websocket.Conn, stream locationv1.LocationService_StreamLocationClient) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			return err
		}
		if err := connection.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
		if err := connection.WriteJSON(memberLocation{
			GroupID: update.GetGroupId(), UserID: update.GetUserId(),
			Latitude: update.GetLatitude(), Longitude: update.GetLongitude(),
			TimestampUnixMs: update.GetTimestampUnixMs(),
		}); err != nil {
			return err
		}
	}
}

func closeLocationConnection(connection *websocket.Conn, err error) {
	closeCode, reason := websocket.CloseNormalClosure, "location stream ended"
	if errors.Is(err, io.EOF) {
		// A clean gRPC stream end is a normal WebSocket closure.
	} else if errors.Is(err, context.Canceled) || status.Code(err) == codes.Canceled {
		closeCode, reason = websocket.CloseGoingAway, "location stream canceled"
	} else if errors.Is(err, websocket.ErrReadLimit) {
		closeCode, reason = websocket.CloseMessageTooBig, "location message too large"
	} else if errors.Is(err, errInvalidCoordinates) {
		closeCode, reason = websocket.CloseInvalidFramePayloadData, "invalid coordinates"
	} else if status.Code(err) == codes.PermissionDenied || status.Code(err) == codes.Unauthenticated {
		closeCode, reason = websocket.ClosePolicyViolation, "location access denied"
	} else if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		closeCode, reason = websocket.CloseInternalServerErr, "location stream failed"
	}
	_ = connection.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, reason), time.Now().Add(time.Second))
}
