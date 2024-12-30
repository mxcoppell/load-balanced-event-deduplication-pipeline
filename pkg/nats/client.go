package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	StreamName = "KEEPALIVE_EXPIRATION"
	Subject    = "Stream.Keepalive.Expiration.*"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
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

	return &Client{
		nc: nc,
		js: js,
	}, nil
}

// PublishExpiredKey publishes an expired key event to the stream
func (c *Client) PublishExpiredKey(ctx context.Context, key string) error {
	msg := &nats.Msg{
		Subject: Subject,
		Data:    []byte(key),
	}

	_, err := c.js.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// SubscribeExpiredKeys subscribes to expired key events
func (c *Client) SubscribeExpiredKeys(ctx context.Context, handler func(key string)) error {
	_, err := c.js.Subscribe(Subject, func(msg *nats.Msg) {
		handler(string(msg.Data))
		msg.Ack()
	}, nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	return nil
}

// Close closes the NATS connection
func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
