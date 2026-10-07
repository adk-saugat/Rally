package handler

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	locationv1 "github.com/adk-saugat/rally/gen/location/v1"
	"github.com/adk-saugat/rally/services/location/internal/service"
	"github.com/adk-saugat/rally/shared/types"
	"github.com/adk-saugat/rally/shared/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocationHandler struct {
	locationv1.UnimplementedLocationServiceServer
	locations service.LocationService
}

func NewLocationHandler(locations service.LocationService) *LocationHandler {
	return &LocationHandler{locations: locations}
}

func (handler *LocationHandler) StreamLocation(stream locationv1.LocationService_StreamLocationServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	userID := utils.UserIDFromContext(ctx)
	if userID == "" {
		return types.GRPCError(types.ErrUnauthorized)
	}

	firstRequest, err := stream.Recv()
	if err != nil {
		return recvError(err)
	}

	firstUpdate, err := updateFromRequest(firstRequest, userID)
	if err != nil {
		return types.GRPCError(err)
	}

	if err := handler.locations.Join(ctx, firstUpdate.GroupID, userID); err != nil {
		return mapServiceError(err)
	}

	updates, err := handler.locations.Subscribe(ctx, firstUpdate.GroupID, userID)
	if err != nil {
		return mapServiceError(err)
	}

	if err := handler.locations.Publish(ctx, firstUpdate); err != nil {
		return mapServiceError(err)
	}

	recvDone := make(chan error, 1)
	go func() {
		recvDone <- handler.receiveUpdates(ctx, stream, firstUpdate.GroupID, userID)
	}()

	for {
		select {
		case err := <-recvDone:
			return err
		case update, ok := <-updates:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return status.Error(codes.Unavailable, "location subscription ended")
			}
			if err := stream.Send(requestFromUpdate(update)); err != nil {
				return err
			}
		}
	}
}

func (handler *LocationHandler) receiveUpdates(ctx context.Context, stream locationv1.LocationService_StreamLocationServer, groupID, userID string) error {
	for {
		request, err := stream.Recv()
		if err != nil {
			return recvError(err)
		}

		update, err := updateFromRequest(request, userID)
		if err != nil {
			return types.GRPCError(err)
		}
		if update.GroupID != groupID {
			return types.GRPCError(types.ErrInvalidInput)
		}

		if err := handler.locations.Publish(ctx, update); err != nil {
			return mapServiceError(err)
		}
	}
}

func requestFromUpdate(update service.Update) *locationv1.LocationUpdate {
	return &locationv1.LocationUpdate{
		GroupId:         update.GroupID,
		UserId:          update.UserID,
		Latitude:        update.Latitude,
		Longitude:       update.Longitude,
		TimestampUnixMs: update.TimestampUnixMs,
	}
}

func updateFromRequest(request *locationv1.LocationUpdate, userID string) (service.Update, error) {
	if request == nil {
		return service.Update{}, types.ErrInvalidInput
	}

	groupID := strings.TrimSpace(request.GetGroupId())
	if groupID == "" {
		return service.Update{}, types.ErrInvalidInput
	}

	timestampUnixMs := request.GetTimestampUnixMs()
	if timestampUnixMs == 0 {
		timestampUnixMs = time.Now().UnixMilli()
	}

	return service.Update{
		GroupID:         groupID,
		UserID:          userID,
		Latitude:        request.GetLatitude(),
		Longitude:       request.GetLongitude(),
		TimestampUnixMs: timestampUnixMs,
	}, nil
}

func recvError(err error) error {
	if err == nil || errors.Is(err, io.EOF) {
		return nil
	}
	return mapServiceError(err)
}

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok && status.Code(err) != codes.Unknown {
		return err
	}
	return types.GRPCError(err)
}
