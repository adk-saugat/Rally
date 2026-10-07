package service

import (
	"context"
	"strings"

	"github.com/adk-saugat/rally/services/location/internal/repository"
	"github.com/adk-saugat/rally/shared/types"
)

type Update struct {
	GroupID         string
	UserID          string
	Latitude        float64
	Longitude       float64
	TimestampUnixMs int64
}

type GroupDirectory interface {
	IsMember(ctx context.Context, groupID string) (bool, error)
}

type LocationService interface {
	Join(ctx context.Context, groupID, userID string) error
	Publish(ctx context.Context, update Update) error
	Subscribe(ctx context.Context, groupID, userID string) (<-chan Update, error)
}

type locationService struct {
	positions repository.PositionRepository
	broker    repository.LocationBroker
	groups    GroupDirectory
}

func NewLocationService(positions repository.PositionRepository, broker repository.LocationBroker, groups GroupDirectory) LocationService {
	return &locationService{
		positions: positions,
		broker:    broker,
		groups:    groups,
	}
}

func (service *locationService) Join(ctx context.Context, groupID, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return types.ErrUnauthorized
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return types.ErrInvalidInput
	}

	isMember, err := service.groups.IsMember(ctx, groupID)
	if err != nil {
		return err
	}
	if !isMember {
		return types.ErrPermissionDenied
	}
	return nil
}

func (service *locationService) Publish(ctx context.Context, update Update) error {
	if err := validateUpdate(update); err != nil {
		return err
	}

	err := service.positions.Set(ctx, update.GroupID, update.UserID, repository.Position{
		Latitude:        update.Latitude,
		Longitude:       update.Longitude,
		TimestampUnixMs: update.TimestampUnixMs,
	})
	if err != nil {
		return err
	}

	return service.broker.Publish(ctx, repository.LocationEvent{
		GroupID:         update.GroupID,
		UserID:          update.UserID,
		Latitude:        update.Latitude,
		Longitude:       update.Longitude,
		TimestampUnixMs: update.TimestampUnixMs,
	})
}

// Subscribe streams location updates from other members of the group.
// The returned channel is closed when ctx is cancelled.
func (service *locationService) Subscribe(ctx context.Context, groupID, userID string) (<-chan Update, error) {
	events, err := service.broker.Subscribe(ctx, groupID)
	if err != nil {
		return nil, err
	}

	updates := make(chan Update)
	go func() {
		defer close(updates)
		for event := range events {
			if event.UserID == userID {
				continue
			}

			update := Update{
				GroupID:         event.GroupID,
				UserID:          event.UserID,
				Latitude:        event.Latitude,
				Longitude:       event.Longitude,
				TimestampUnixMs: event.TimestampUnixMs,
			}
			select {
			case updates <- update:
			case <-ctx.Done():
				return
			}
		}
	}()

	return updates, nil
}

func validateUpdate(update Update) error {
	if strings.TrimSpace(update.GroupID) == "" || strings.TrimSpace(update.UserID) == "" {
		return types.ErrInvalidInput
	}
	if update.Latitude < -90 || update.Latitude > 90 {
		return types.ErrInvalidInput
	}
	if update.Longitude < -180 || update.Longitude > 180 {
		return types.ErrInvalidInput
	}
	return nil
}
