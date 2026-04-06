// Package pubsub provides a Redis Pub/Sub subscriber with automatic reconnection.
// Used for cache invalidation, feature flag sync, and cross-instance events.
package pubsub

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// MessageHandler processes a message received from a channel.
type MessageHandler func(channel, payload string)

// Subscriber wraps Redis Pub/Sub with auto-reconnection.
type Subscriber struct {
	client     *goredis.Client
	maxBackoff time.Duration
}

// NewSubscriber creates a Pub/Sub subscriber.
func NewSubscriber(client *goredis.Client) *Subscriber {
	return &Subscriber{
		client:     client,
		maxBackoff: 30 * time.Second,
	}
}

// Subscribe listens to a channel and calls handler for each message.
// Blocks until ctx is cancelled. Automatically reconnects on failure
// with exponential backoff (max 30s).
func (s *Subscriber) Subscribe(ctx context.Context, channel string, handler MessageHandler) {
	s.listen(ctx, func() *goredis.PubSub {
		return s.client.Subscribe(ctx, channel)
	}, handler, channel)
}

// PSubscribe listens using a glob pattern (e.g. "xbank.*").
func (s *Subscriber) PSubscribe(ctx context.Context, pattern string, handler MessageHandler) {
	s.listen(ctx, func() *goredis.PubSub {
		return s.client.PSubscribe(ctx, pattern)
	}, handler, pattern)
}

// Publish sends a message to a channel.
func (s *Subscriber) Publish(ctx context.Context, channel, message string) error {
	return s.client.Publish(ctx, channel, message).Err()
}

func (s *Subscriber) listen(ctx context.Context, subFn func() *goredis.PubSub, handler MessageHandler, name string) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		pubsub := subFn()
		ch := pubsub.Channel()

		logger.Log.Info("pubsub: subscribed", zap.String("channel", name))
		backoff = time.Second // reset on successful connect

		for {
			select {
			case <-ctx.Done():
				pubsub.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					// Channel closed — reconnect
					logger.Log.Warn("pubsub: channel closed, reconnecting",
						zap.String("channel", name),
					)
					pubsub.Close()
					goto reconnect
				}
				handler(msg.Channel, msg.Payload)
			}
		}

	reconnect:
		if backoff < s.maxBackoff {
			backoff *= 2
		}
		logger.Log.Info("pubsub: reconnecting",
			zap.String("channel", name),
			zap.Duration("backoff", backoff),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
	}
}
