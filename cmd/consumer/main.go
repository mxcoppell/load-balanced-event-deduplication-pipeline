package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mxie/load-balanced-event-deduplication-pipeline/pkg/nats"
	"github.com/mxie/load-balanced-event-deduplication-pipeline/pkg/redis"
)

const (
	defaultRedisAddr = "redis:6379"
	defaultNatsURL   = "nats://nats:4222"
	defaultDedupTTL  = 5 * time.Second
)

func main() {
	// Get pod name for consumer ID
	consumerID := os.Getenv("HOSTNAME")
	if consumerID == "" {
		consumerID = "unknown"
	}
	log.Printf("Starting consumer with ID: %s", consumerID)

	// Create Redis client
	redisClient, err := redis.NewClient(defaultRedisAddr)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Create NATS client
	natsClient, err := nats.NewClient(defaultNatsURL)
	if err != nil {
		log.Fatalf("Failed to create NATS client: %v", err)
	}
	defer natsClient.Close()

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

	// Subscribe to NATS stream for deduplicated events
	if err := natsClient.SubscribeExpiredKeys(ctx, func(key string) {
		log.Printf("Consumer %s processing deduplicated key: %s", consumerID, key)
		// Increment consumer-specific metric
		if err := redisClient.IncrementConsumerMetric(ctx, consumerID); err != nil {
			log.Printf("Failed to increment consumer metric: %v", err)
		}
		// Increment consumed metric
		if err := redisClient.IncrementConsumed(ctx); err != nil {
			log.Printf("Failed to increment consumed metric: %v", err)
		}
	}); err != nil {
		log.Fatalf("Failed to subscribe to NATS stream: %v", err)
	}

	// Process Redis expired keys with automatic reconnection
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("Subscribing to Redis key expiration events...")
			// Subscribe to Redis key expiration events
			psc := redisClient.Subscribe(ctx, "__keyevent@0__:expired")

			// Process messages until error or context cancellation
			for {
				select {
				case <-ctx.Done():
					psc.Close()
					return
				default:
					msg, err := psc.ReceiveMessage(ctx)
					if err != nil {
						log.Printf("Error receiving message: %v, will reconnect in 5 seconds", err)
						psc.Close()
						time.Sleep(5 * time.Second)
						break
					}

					// Process the expired key
					go handleRedisExpiredKey(ctx, msg.Payload, redisClient, natsClient, consumerID)
				}
			}
		}
	}
}

// handleRedisExpiredKey handles Redis key expiration events and publishes them to NATS
func handleRedisExpiredKey(ctx context.Context, key string, redisClient *redis.Client, natsClient *nats.Client, consumerID string) {
	log.Printf("Consumer %s received Redis expired key: %s", consumerID, key)

	// Ignore dedup keys
	if strings.HasPrefix(key, redis.DedupPrefix) {
		log.Printf("Ignoring dedup key: %s", key)
		return
	}

	// Try to create dedup key
	ok, err := redisClient.CreateDedupKey(ctx, key, defaultDedupTTL)
	if err != nil {
		log.Printf("Failed to create dedup key for %s: %v", key, err)
		return
	}

	// If dedup key already exists, ignore this event
	if !ok {
		log.Printf("Dedup key already exists for %s, ignoring", key)
		return
	}

	// Publish expired key to NATS
	if err := natsClient.PublishExpiredKey(ctx, key); err != nil {
		log.Printf("Failed to publish key %s to NATS: %v", key, err)
		return
	}
	log.Printf("Successfully published key %s to NATS", key)
}
