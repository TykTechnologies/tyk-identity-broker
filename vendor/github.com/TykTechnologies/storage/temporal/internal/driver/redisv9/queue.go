package redisv9

import (
	"context"
	"errors"
	"time"

	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
	"github.com/redis/go-redis/v9"
)

// subscribeAdapter is an adapter for redis.PubSub to satisfy model.Subscription interface.
// Receive() method returns a model.Message instead of an interface{}.
type subscriptionAdapter struct {
	pubSub *redis.PubSub
}

// messageAdapter is an adapter to satisfy model.Message interface.
// Channel() and Payload() methods return the channel and payload of the message.
// Type() method returns the type of the message.
type messageAdapter struct {
	msg interface{}
}

// newSubscriptionAdapter returns a new subscriptionAdapter.
func newSubscriptionAdapter(pubSub *redis.PubSub) *subscriptionAdapter {
	return &subscriptionAdapter{pubSub: pubSub}
}

// newMessageAdapter returns a new messageAdapter.
func newMessageAdapter(msg interface{}) *messageAdapter {
	return &messageAdapter{msg: msg}
}

// Type returns the message type.
func (m *messageAdapter) Type() string {
	switch m.msg.(type) {
	case *redis.Message:
		return model.MessageTypeMessage
	case *redis.Subscription:
		return model.MessageTypeSubscription
	default:
		return temperr.UnknownMessageType.Error()
	}
}

// Channel returns the channel the message was received on.
func (m *messageAdapter) Channel() (string, error) {
	switch msg := m.msg.(type) {
	case *redis.Message:
		return msg.Channel, nil
	case *redis.Subscription:
		return msg.Channel, nil
	default:
		return "", temperr.UnknownMessageType
	}
}

// Payload returns the message payload.
func (m *messageAdapter) Payload() (string, error) {
	switch msg := m.msg.(type) {
	case *redis.Message:
		return msg.Payload, nil
	case *redis.Subscription:
		return msg.Kind, nil
	default:
		return "", temperr.UnknownMessageType
	}
}

// Receive waits for and returns the next message from the subscription.
func (r *subscriptionAdapter) Receive(ctx context.Context) (model.Message, error) {
	timeout := time.Duration(0)
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	msg, err := r.pubSub.ReceiveTimeout(ctx, timeout)
	if err != nil {
		if errors.Is(err, redis.ErrClosed) {
			return nil, temperr.ClosedConnection
		}

		return nil, err
	}

	return newMessageAdapter(msg), nil
}

// Close closes the subscription and cleans up resources.
func (r *subscriptionAdapter) Close() error {
	return r.pubSub.Close()
}

// Publish sends a message to the specified channel.
func (r *RedisV9) Publish(ctx context.Context, channel, message string) (int64, error) {
	res, err := r.client.Publish(ctx, channel, message).Result()
	if err != nil {
		if errors.Is(err, redis.ErrClosed) {
			return 0, temperr.ClosedConnection
		}
	}

	return res, err
}

// Subscribe initializes a subscription to one or more channels.
func (r *RedisV9) Subscribe(ctx context.Context, channels ...string) model.Subscription {
	sub := r.client.Subscribe(ctx, channels...)

	adapterSub := newSubscriptionAdapter(sub)

	return adapterSub
}
