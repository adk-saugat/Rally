package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/adk-saugat/rally/gateway/internal/clients"
	"github.com/adk-saugat/rally/gateway/internal/config"
	"github.com/adk-saugat/rally/gateway/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	authClient, err := clients.NewAuthClient(cfg.AuthAddress)
	if err != nil {
		log.Fatalf("connect auth: %v", err)
	}
	defer authClient.Close()

	groupClient, err := clients.NewGroupClient(cfg.GroupAddress)
	if err != nil {
		log.Fatalf("connect group: %v", err)
	}
	defer groupClient.Close()

	locationClient, err := clients.NewLocationClient(cfg.LocationAddress)
	if err != nil {
		log.Fatalf("connect location: %v", err)
	}
	defer locationClient.Close()

	authHandler := handler.NewAuthHandler(authClient)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", handler.Health)
	router.POST("/auth/signup", authHandler.SignUp)
	router.POST("/auth/login", authHandler.Login)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("gateway listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}
