package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	AuthAddress     string
	GroupAddress    string
	LocationAddress string
}

func Load() Config {
	_ = godotenv.Load()
	_ = godotenv.Load("gateway/.env")

	return Config{
		Port:            env("PORT", "8080"),
		AuthAddress:     env("AUTH_ADDRESS", "localhost:50051"),
		GroupAddress:    env("GROUP_ADDRESS", "localhost:50052"),
		LocationAddress: env("LOCATION_ADDRESS", "localhost:50053"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
