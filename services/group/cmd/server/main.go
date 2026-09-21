package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	groupv1 "github.com/adk-saugat/rally/gen/group/v1"
	"github.com/adk-saugat/rally/services/group/internal/config"
	"github.com/adk-saugat/rally/services/group/internal/handler"
	"github.com/adk-saugat/rally/services/group/internal/repository"
	"github.com/adk-saugat/rally/services/group/internal/service"
	"github.com/adk-saugat/rally/services/group/internal/utils"
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

	groups := repository.NewGroupRepository(pool)
	groupService := service.NewGroupService(groups)
	groupHandler := handler.NewGroupHandler(groupService)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(utils.UnaryAuth(cfg.JWTSecret)))
	groupv1.RegisterGroupServiceServer(grpcServer, groupHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	log.Printf("group listening on %s", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
