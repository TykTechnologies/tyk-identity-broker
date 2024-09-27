package model

import (
	"context"
	"time"
)

const (
	RedisV9Type = "redisv9"
)

type Connector interface {
	// Disconnect disconnects from the backend
	Disconnect(context.Context) error

	// Ping executes a ping to the backend
	Ping(context.Context) error

	// Type returns the  connector type
	Type() string

	// As converts i to driver-specific types.
	// Same concept as https://gocloud.dev/concepts/as/ but for connectors.
	As(i interface{}) bool
}

type List interface {
	// Remove the first count occurrences of elements equal to element from the list stored at key.
	Remove(ctx context.Context, key string, count int64, element interface{}) (int64, error)

	// Returns the specified elements of the list stored at key.
	// The offsets start and stop are zero-based indexes.
	Range(ctx context.Context, key string, start, stop int64) ([]string, error)

	// Returns the length of the list stored at key.
	Length(ctx context.Context, key string) (int64, error)

	// Insert all the specified values at the head of the list stored at key.
	// If key does not exist, it is created.
	// pipelined: If true, the operation is pipelined and executed in a single roundtrip.
	Prepend(ctx context.Context, pipelined bool, key string, values ...[]byte) error

	// Insert all the specified values at the tail of the list stored at key.
	// If key does not exist, it is created.
	// pipelined: If true, the operation is pipelined and executed in a single roundtrip.
	Append(ctx context.Context, pipelined bool, key string, values ...[]byte) error

	// Pop removes and returns the first count elements of the list stored at key.
	// If stop is -1, all the elements from start to the end of the list are removed and returned.
	Pop(ctx context.Context, key string, stop int64) ([]string, error)
}

type KeyValue interface {
	// Get retrieves the value for a given key
	Get(ctx context.Context, key string) (value string, err error)
	// Set sets the string value of a key
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	// SetIfNotExist sets the string value of a key if the key does not exist.
	// Returns true if the key was set, false otherwise.
	SetIfNotExist(ctx context.Context, key, value string, expiration time.Duration) (bool, error)
	// Delete removes the specified keys
	Delete(ctx context.Context, key string) error
	// Increment atomically increments the integer value of a key by one
	Increment(ctx context.Context, key string) (newValue int64, err error)
	// Decrement atomically decrements the integer value of a key by one
	Decrement(ctx context.Context, key string) (newValue int64, err error)
	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (exists bool, err error)
	// Expire sets a timeout on key. After the timeout has expired, the key will automatically be deleted
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// TTL returns the remaining time to live of a key that has a timeout
	TTL(ctx context.Context, key string) (ttl int64, err error)
	// DeleteKeys deletes all keys that match the given pattern
	DeleteKeys(ctx context.Context, keys []string) (numberOfDeletedKeys int64, err error)
	// DeleteScanMatch deletes all keys that match the given pattern
	DeleteScanMatch(ctx context.Context, pattern string) (numberOfDeletedKeys int64, err error)
	// Keys returns all keys that match the given pattern
	Keys(ctx context.Context, pattern string) (keys []string, err error)
	// GetMulti returns the values of all specified keys
	GetMulti(ctx context.Context, keys []string) (values []interface{}, err error)
	// GetKeysAndValuesWithFilter returns all keys and values that match the given pattern
	GetKeysAndValuesWithFilter(ctx context.Context, pattern string) (keysAndValues map[string]interface{}, err error)
	// GetKeysWithOpts retrieves keys with options like filter, cursor, and count
	GetKeysWithOpts(ctx context.Context, searchStr string, cursors map[string]uint64,
		count int64) (keys []string, updatedCursor map[string]uint64, continueScan bool, err error)
}

type Flusher interface {
	// FlushAll deletes all keys the database
	FlushAll(ctx context.Context) error
}

type SortedSet interface {
	// AddScoredMember adds a member with a specific score to a sorted set.
	// Returns the number of elements added to the sorted set.
	AddScoredMember(ctx context.Context, key, member string, score float64) (int64, error)

	// GetMembersByScoreRange retrieves members and their scores from a sorted set
	// within the given score range.
	// Returns slices of members and their scores, and an error if any.
	GetMembersByScoreRange(ctx context.Context, key, minScore, maxScore string) ([]interface{}, []float64, error)

	// RemoveMembersByScoreRange removes members from a sorted set within a specified score range.
	// Returns the number of members removed.
	RemoveMembersByScoreRange(ctx context.Context, key, minScore, maxScore string) (int64, error)
}

type Set interface {
	// Returns all the members of the set value stored at key.
	Members(ctx context.Context, key string) ([]string, error)

	// Add the specified members to the set stored at key.
	// Specified members that are already a member of this set are ignored.
	// If key does not exist, a new set is created before adding the specified members.
	AddMember(ctx context.Context, key, member string) error

	// Remove the specified members from the set stored at key.
	// Specified members that are not a member of this set are ignored.
	RemoveMember(ctx context.Context, key, member string) error

	// Returns if member is a member of the set stored at key.
	IsMember(ctx context.Context, key, member string) (bool, error)
}

// Queue interface represents a pub/sub queue with methods to publish messages
// and subscribe to channels.
type Queue interface {
	// Publish sends a message to the specified channel.
	// It returns the number of clients that received the message.
	Publish(ctx context.Context, channel, message string) (int64, error)

	// Subscribe initializes a subscription to one or more channels.
	// It returns a Subscription interface that allows receiving messages and closing the subscription.
	Subscribe(ctx context.Context, channels ...string) Subscription
}

// Subscription interface represents a subscription to one or more channels.
// It allows receiving messages and closing the subscription.
type Subscription interface {
	// Receive waits for and returns the next message from the subscription.
	Receive(ctx context.Context) (Message, error)

	// Close closes the subscription and cleans up resources.
	Close() error
}

// Message represents a message received from a subscription.
type Message interface {
	// Type returns the message type.
	// It can be one of the following:
	// - "message": a message received from a subscription with a payload and channel
	// - "subscription": a subscription confirmation message with a channel
	Type() string
	// Channel returns the channel the message was received on.
	// It can be one of the following depending on the message type:
	// - the channel the message was received on
	// - the channel the subscription was created on
	// - an empty string, returning an error
	Channel() (string, error)
	// Payload returns the message payload.
	// It can be one of the following depending on the message type:
	// - the message payload
	// - the subscription kind (e.g. "subscribe", "unsubscribe")
	// - an empty string, returning an error
	Payload() (string, error)
}
