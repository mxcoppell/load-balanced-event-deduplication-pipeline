package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	KeyPrefix        = "gen-key:"
	DedupPrefix      = "dedup:"
	MetricsGenerated = "metrics:generated"
	MetricsConsumed  = "metrics:consumed"
	MetricsConsumer  = "metrics:consumer:" // Prefix for per-consumer metrics
)

type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client
func NewClient(addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Subscribe subscribes to Redis Pub/Sub channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// GenerateKey creates a new key with expiration
func (c *Client) GenerateKey(ctx context.Context, seqNum int64, ttl time.Duration) error {
	key := fmt.Sprintf("%s%d", KeyPrefix, seqNum)
	if err := c.rdb.Set(ctx, key, seqNum, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	// Increment generated keys metric
	if err := c.rdb.Incr(ctx, MetricsGenerated).Err(); err != nil {
		return fmt.Errorf("failed to increment generated metric: %w", err)
	}

	return nil
}

// CreateDedupKey creates a deduplication key if it doesn't exist
func (c *Client) CreateDedupKey(ctx context.Context, originalKey string, ttl time.Duration) (bool, error) {
	dedupKey := DedupPrefix + originalKey

	// Try to set dedup key only if it doesn't exist
	ok, err := c.rdb.SetNX(ctx, dedupKey, 1, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set dedup key %s: %w", dedupKey, err)
	}

	return ok, nil
}

// IncrementConsumed increments the consumed keys metric
func (c *Client) IncrementConsumed(ctx context.Context) error {
	if err := c.rdb.Incr(ctx, MetricsConsumed).Err(); err != nil {
		return fmt.Errorf("failed to increment consumed metric: %w", err)
	}
	return nil
}

// GetMetrics returns the current metrics
func (c *Client) GetMetrics(ctx context.Context) (generated, consumed int64, err error) {
	pipe := c.rdb.Pipeline()
	genCmd := pipe.Get(ctx, MetricsGenerated)
	consCmd := pipe.Get(ctx, MetricsConsumed)

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return 0, 0, fmt.Errorf("failed to get metrics: %w", err)
	}

	generated, _ = genCmd.Int64()
	consumed, _ = consCmd.Int64()
	return generated, consumed, nil
}

// ResetMetrics resets all metrics to zero
func (c *Client) ResetMetrics(ctx context.Context) error {
	// Get all consumer metric keys
	pattern := MetricsConsumer + "*"
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get consumer metrics keys: %w", err)
	}

	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, MetricsGenerated, 0, 0)
	pipe.Set(ctx, MetricsConsumed, 0, 0)

	// Delete all consumer metrics
	if len(keys) > 0 {
		pipe.Del(ctx, keys...)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to reset metrics: %w", err)
	}
	return nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}

// IncrementConsumerMetric increments the metric for a specific consumer
func (c *Client) IncrementConsumerMetric(ctx context.Context, consumerID string) error {
	key := MetricsConsumer + consumerID
	if err := c.rdb.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to increment consumer metric for %s: %w", consumerID, err)
	}
	return nil
}

// GetConsumerMetrics returns a map of consumer IDs to their processed event counts
func (c *Client) GetConsumerMetrics(ctx context.Context) (map[string]int64, error) {
	pattern := MetricsConsumer + "*"
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer metrics keys: %w", err)
	}

	metrics := make(map[string]int64)
	if len(keys) == 0 {
		return metrics, nil
	}

	// Get all values in a single pipeline
	pipe := c.rdb.Pipeline()
	cmds := make(map[string]*redis.StringCmd)
	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get consumer metrics values: %w", err)
	}

	for key, cmd := range cmds {
		val, _ := cmd.Int64()
		consumerID := strings.TrimPrefix(key, MetricsConsumer)
		metrics[consumerID] = val
	}

	return metrics, nil
}
