package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	StreamName = "WORKGROUPPOLICY"
	Subject    = "Stream.Workgroup.Policy.Events"
	QueueGroup = "key_expiration_processors"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// initStream creates the stream if it doesn't exist
func (c *Client) initStream() error {
	// Check if stream exists
	stream, err := c.js.StreamInfo(StreamName)
	if err == nil && stream != nil {
		return nil // Stream already exists
	}

	// Create stream with WorkQueue retention
	_, err = c.js.AddStream(&nats.StreamConfig{
		Name:      StreamName,
		Subjects:  []string{Subject},
		Retention: nats.WorkQueuePolicy,
		Storage:   nats.MemoryStorage,
		MaxAge:    24 * time.Hour,
		Replicas:  1,
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	return nil
}

// NewClient creates a new NATS client with JetStream enabled
func NewClient(url string) (*Client, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	client := &Client{
		nc: nc,
		js: js,
	}

	// Initialize stream
	if err := client.initStream(); err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to initialize stream: %w", err)
	}

	return client, nil
}

// PublishExpiredKey publishes an expired key event to the stream
func (c *Client) PublishExpiredKey(ctx context.Context, key string) error {
	msg := &nats.Msg{
		Subject: Subject,
		Data:    []byte(key),
	}

	// Publish with context for timeout control
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := c.js.PublishMsg(msg)
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
		return nil
	}
}

// SubscribeExpiredKeys subscribes to expired key events using queue groups for even distribution
func (c *Client) SubscribeExpiredKeys(ctx context.Context, handler func(key string)) error {
	// Create a consumer with queue group for even distribution
	_, err := c.js.QueueSubscribe(
		Subject,
		QueueGroup,
		func(msg *nats.Msg) {
			// Process the message
			handler(string(msg.Data))
			// Acknowledge successful processing
			msg.Ack()
		},
		// Configure subscription options
		nats.ManualAck(),            // Enable manual acknowledgment
		nats.AckWait(5*time.Second), // Set acknowledgment timeout
		nats.MaxDeliver(3),          // Maximum redelivery attempts
		nats.DeliverAll(),           // Deliver all messages in the stream
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Monitor context for cancellation
	go func() {
		<-ctx.Done()
		c.Close()
	}()

	return nil
}

// Close closes the NATS connection
func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
