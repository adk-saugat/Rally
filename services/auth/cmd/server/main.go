package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	authv1 "github.com/adk-saugat/rally/gen/auth/v1"
	"github.com/adk-saugat/rally/services/auth/internal/config"
	"github.com/adk-saugat/rally/services/auth/internal/handler"
	"github.com/adk-saugat/rally/services/auth/internal/repository"
	"github.com/adk-saugat/rally/services/auth/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	users := repository.NewUserRepository(pool)
	auth := service.NewAuthService(users, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(auth)

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	log.Printf("auth listening on %s", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
