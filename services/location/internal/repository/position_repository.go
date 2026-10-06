package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Position struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TimestampUnixMs int64   `json:"timestamp_unix_ms"`
}

type PositionRepository interface {
	Set(ctx context.Context, groupID, userID string, position Position) error
	Get(ctx context.Context, groupID, userID string) (*Position, error)
}

type positionRepository struct {
	client *redis.Client
}

func NewPositionRepository(client *redis.Client) PositionRepository {
	return &positionRepository{client: client}
}

func Connect(ctx context.Context, redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (repo *positionRepository) Set(ctx context.Context, groupID, userID string, position Position) error {
	payload, err := json.Marshal(position)
	if err != nil {
		return err
	}
	return repo.client.Set(ctx, positionKey(groupID, userID), payload, 0).Err()
}

func (repo *positionRepository) Get(ctx context.Context, groupID, userID string) (*Position, error) {
	payload, err := repo.client.Get(ctx, positionKey(groupID, userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var position Position
	if err := json.Unmarshal(payload, &position); err != nil {
		return nil, err
	}
	return &position, nil
}

func positionKey(groupID, userID string) string {
	return fmt.Sprintf("group:%s:user:%s", groupID, userID)
}
