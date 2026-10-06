package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const locationChannelFormat = "group:%s:locations"

type LocationEvent struct {
	GroupID         string  `json:"group_id"`
	UserID          string  `json:"user_id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TimestampUnixMs int64   `json:"timestamp_unix_ms"`
}

type LocationBroker interface {
	Publish(ctx context.Context, event LocationEvent) error
	Subscribe(ctx context.Context, groupID string) (<-chan LocationEvent, error)
}

type locationBroker struct {
	client *redis.Client
}

func NewLocationBroker(client *redis.Client) LocationBroker {
	return &locationBroker{client: client}
}

func (broker *locationBroker) Publish(ctx context.Context, event LocationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf(locationChannelFormat, event.GroupID)
	return broker.client.Publish(ctx, channel, payload).Err()
}

// Subscribe returns a channel of events for the group. The channel is closed
// once ctx is cancelled or the underlying subscription ends.
func (broker *locationBroker) Subscribe(ctx context.Context, groupID string) (<-chan LocationEvent, error) {
	channel := fmt.Sprintf(locationChannelFormat, groupID)
	pubsub := broker.client.Subscribe(ctx, channel)

	// Wait for the subscription to be confirmed so no events published
	// right after Subscribe returns are missed.
	if _, err := pubsub.Receive(ctx); err != nil {
		pubsub.Close()
		return nil, err
	}

	// Must be obtained before Close can run, or the range below never ends.
	messages := pubsub.Channel()
	go func() {
		<-ctx.Done()
		pubsub.Close()
	}()

	events := make(chan LocationEvent)
	go func() {
		defer close(events)
		for message := range messages {
			var event LocationEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				continue
			}

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}
