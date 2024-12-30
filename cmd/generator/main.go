package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mxie/load-balanced-event-deduplication-pipeline/pkg/redis"
)

const (
	defaultRedisAddr = "redis:6379"
)

func main() {
	// Create Redis client
	redisClient, err := redis.NewClient(defaultRedisAddr)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start the web server
	if err := startWebServer(ctx, redisClient); err != nil {
		log.Fatalf("Web server error: %v", err)
	}
}
