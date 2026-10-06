package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	locationv1 "github.com/adk-saugat/rally/gen/location/v1"
	"github.com/adk-saugat/rally/services/location/internal/clients"
	"github.com/adk-saugat/rally/services/location/internal/config"
	"github.com/adk-saugat/rally/services/location/internal/handler"
	"github.com/adk-saugat/rally/services/location/internal/repository"
	"github.com/adk-saugat/rally/services/location/internal/service"
	"github.com/adk-saugat/rally/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient, err := repository.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	positions := repository.NewPositionRepository(redisClient)

	groupClient, err := clients.NewGroupClient(cfg.GroupAddress)
	if err != nil {
		log.Fatalf("connect group: %v", err)
	}
	defer groupClient.Close()

	locations := service.NewLocationService(positions, groupClient)
	locationHandler := handler.NewLocationHandler(locations)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(utils.UnaryAuth(cfg.JWTSecret)),
		grpc.StreamInterceptor(utils.StreamAuth(cfg.JWTSecret)),
	)
	locationv1.RegisterLocationServiceServer(grpcServer, locationHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	log.Printf("location listening on %s", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
