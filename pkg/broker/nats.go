// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package broker

import (
	"fmt"
	"time"

	natssdk "github.com/nats-io/nats.go"
)

// Msg is a NATS message.
type Msg = natssdk.Msg

// MsgHandler handles an inbound NATS message.
type MsgHandler = natssdk.MsgHandler

// Subscription is an active NATS subscription.
type Subscription = natssdk.Subscription

// Client wraps the NATS connection.
type Client struct {
	config Config
	conn   *natssdk.Conn
}

// New returns a broker client from app.broker config.
func New() (*Client, error) {
	config := GetConfig()

	switch config.Provider {
	case "nats":
		return NewNatsClient(config)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, config.Provider)
	}
}

// NewNatsClient creates a NATS-backed broker client.
func NewNatsClient(config Config) (*Client, error) {
	conn, err := natssdk.Connect(
		config.NATS.URL,
		natssdk.Name(config.NATS.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	return &Client{
		config: config,
		conn:   conn,
	}, nil
}

// Config returns the broker configuration used by this client.
func (c *Client) Config() Config {
	return c.config
}

// Conn returns the underlying NATS connection.
func (c *Client) Conn() *natssdk.Conn {
	return c.conn
}

// Publish sends data on the given subject.
func (c *Client) Publish(subject string, data []byte) error {
	err := c.conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}

	return nil
}

// Subscribe registers an async handler for subject.
func (c *Client) Subscribe(subject string, handler MsgHandler) (*Subscription, error) {
	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}

	return sub, nil
}

// QueueSubscribe registers an async queue subscriber for subject.
func (c *Client) QueueSubscribe(subject, queue string, handler MsgHandler) (*Subscription, error) {
	sub, err := c.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return nil, fmt.Errorf("nats queue subscribe: %w", err)
	}

	return sub, nil
}

// Request sends a request and waits for a reply.
func (c *Client) Request(subject string, data []byte, timeout time.Duration) (*Msg, error) {
	msg, err := c.conn.Request(subject, data, timeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	return msg, nil
}

// Flush drains pending outbound messages.
func (c *Client) Flush() error {
	err := c.conn.Flush()
	if err != nil {
		return fmt.Errorf("nats flush: %w", err)
	}

	return nil
}

// Ping checks NATS connectivity with a round-trip flush.
func (c *Client) Ping(timeout time.Duration) error {
	err := c.conn.FlushTimeout(timeout)
	if err != nil {
		return fmt.Errorf("nats ping: %w", err)
	}

	return nil
}

// Close closes the NATS connection.
func (c *Client) Close() {
	c.conn.Close()
}
