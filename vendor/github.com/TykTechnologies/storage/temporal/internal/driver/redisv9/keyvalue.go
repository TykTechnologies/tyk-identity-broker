package redisv9

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/TykTechnologies/storage/temporal/temperr"
	"github.com/redis/go-redis/v9"
)

// Get retrieves the value for a given key from Redis
func (r *RedisV9) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", temperr.KeyEmpty
	}

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", temperr.KeyNotFound
		}

		return "", err
	}

	return result, nil
}

// Set sets the string value of a key
func (r *RedisV9) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	if key == "" {
		return temperr.KeyEmpty
	}

	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete removes the specified keys
func (r *RedisV9) Delete(ctx context.Context, key string) error {
	if key == "" {
		return temperr.KeyEmpty
	}

	_, err := r.client.Del(ctx, key).Result()

	return err
}

// Increment atomically increments the integer value of a key by one
func (r *RedisV9) Increment(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, temperr.KeyEmpty
	}

	res, err := r.client.Incr(ctx, key).Result()
	if err != nil && strings.EqualFold(err.Error(), "ERR value is not an integer or out of range") {
		return 0, temperr.KeyMisstype
	}

	return res, err
}

// Decrement atomically decrements the integer value of a key by one
func (r *RedisV9) Decrement(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, temperr.KeyEmpty
	}

	res, err := r.client.Decr(ctx, key).Result()
	if err != nil && strings.EqualFold(err.Error(), "ERR value is not an integer or out of range") {
		return 0, temperr.KeyMisstype
	}

	return res, err
}

// Exists checks if a key exists
func (r *RedisV9) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, temperr.KeyEmpty
	}

	result, err := r.client.Exists(ctx, key).Result()

	return result > 0, err
}

// Expire sets a timeout on key. After the timeout has expired, the key will automatically be deleted
func (r *RedisV9) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if key == "" {
		return temperr.KeyEmpty
	}

	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the remaining time to live of a key that has a timeout
func (r *RedisV9) TTL(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return -2, temperr.KeyEmpty
	}

	duration, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// since redis-go v8.3.1, if there's no expiration or the key doesn't exists,
	// the ttl returned is measured in nanoseconds
	if duration.Nanoseconds() == -1 || duration.Nanoseconds() == -2 {
		return duration.Nanoseconds(), nil
	}

	return int64(duration.Seconds()), nil
}

// DeleteKeys removes the specified keys. A key is ignored if it does not exist
func (r *RedisV9) DeleteKeys(ctx context.Context, keys []string) (int64, error) {
	if len(keys) == 0 {
		return 0, temperr.KeyEmpty
	}

	switch v := r.client.(type) {
	case *redis.ClusterClient:
		return r.deleteKeysCluster(ctx, v, keys)
	case *redis.Client:
		return v.Del(ctx, keys...).Result()
	default:
		return 0, temperr.InvalidRedisClient
	}
}

// deleteKeysCluster removes the specified keys on a cluster
func (r *RedisV9) deleteKeysCluster(ctx context.Context, cluster *redis.ClusterClient, keys []string) (int64, error) {
	var totalDeleted int64

	for _, key := range keys {
		delCmd := redis.NewIntCmd(ctx, "DEL", key)

		// Process the command, which sends it to the appropriate node
		if err := cluster.Process(ctx, delCmd); err != nil {
			return totalDeleted, err
		}

		// Accumulate the count of deleted keys
		deleted, err := delCmd.Result()
		if err != nil {
			return totalDeleted, err
		}

		totalDeleted += deleted
	}

	return totalDeleted, nil
}

// DeleteScanMatch deletes all keys matching the given pattern
func (r *RedisV9) DeleteScanMatch(ctx context.Context, pattern string) (int64, error) {
	var totalDeleted int64
	var mutex sync.Mutex
	var firstError error

	switch client := r.client.(type) {
	case *redis.ClusterClient:
		err := client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			deleted, err := r.deleteScanMatchSingleNode(ctx, client, pattern)
			if err != nil {
				if firstError == nil {
					firstError = err
				}
				return nil // Continue with other nodes
			}

			mutex.Lock()
			totalDeleted += deleted
			mutex.Unlock()

			return nil
		})

		if errors.Is(err, redis.ErrClosed) || errors.Is(firstError, redis.ErrClosed) {
			return totalDeleted, temperr.ClosedConnection
		}

		if firstError != nil {
			return totalDeleted, firstError
		}
		if err != nil {
			return totalDeleted, err
		}

	case *redis.Client:
		var err error
		totalDeleted, err = r.deleteScanMatchSingleNode(ctx, client, pattern)
		if err != nil {
			if errors.Is(err, redis.ErrClosed) {
				return totalDeleted, temperr.ClosedConnection
			}

			return totalDeleted, err
		}

	default:
		return totalDeleted, temperr.InvalidRedisClient
	}

	return totalDeleted, nil
}

// deleteScanMatchSingleNode deletes all keys matching the given pattern on a single node
func (r *RedisV9) deleteScanMatchSingleNode(ctx context.Context, client redis.Cmdable, pattern string) (int64, error) {
	var deleted, cursor uint64
	var err error

	var keys []string
	keys, _, err = client.Scan(ctx, cursor, pattern, 0).Result()
	if err != nil {
		return int64(deleted), err
	}

	if len(keys) > 0 {
		del, err := client.Del(ctx, keys...).Result()
		if err != nil {
			return int64(deleted), err
		}

		deleted += uint64(del)
	}

	return int64(deleted), nil
}

// Keys returns all keys matching the given pattern
func (r *RedisV9) Keys(ctx context.Context, pattern string) ([]string, error) {
	var sessions []string
	var mutex sync.Mutex
	var firstError error

	switch client := r.client.(type) {
	case *redis.ClusterClient:
		err := client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			keys, err := fetchAllKeys(ctx, client, pattern)
			if err != nil {
				if firstError == nil {
					firstError = err
				}
				return nil // continue with other nodes
			}

			mutex.Lock()
			sessions = append(sessions, keys...)
			mutex.Unlock()

			return nil
		})

		if errors.Is(err, redis.ErrClosed) || errors.Is(firstError, redis.ErrClosed) {
			return nil, temperr.ClosedConnection
		}

		if firstError != nil {
			return nil, firstError
		}
		if err != nil {
			return nil, err
		}

	case *redis.Client:
		keys, err := fetchAllKeys(ctx, client, pattern)
		if err != nil {
			if errors.Is(err, redis.ErrClosed) {
				return nil, temperr.ClosedConnection
			}

			return nil, err
		}

		sessions = keys
	default:
		return nil, temperr.InvalidRedisClient
	}

	return sessions, nil
}

// GetMulti returns the values of all specified keys
func (r *RedisV9) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	switch client := r.client.(type) {
	case *redis.ClusterClient:
		return r.getMultiCluster(ctx, client, keys)
	case *redis.Client:
		return r.getMultiStandalone(ctx, client, keys)
	default:
		return nil, temperr.InvalidRedisClient
	}
}

func (r *RedisV9) getMultiCluster(ctx context.Context,
	client *redis.ClusterClient,
	keys []string,
) ([]interface{}, error) {
	values := make([]interface{}, len(keys))

	for i, key := range keys {
		value, err := r.getValueFromCluster(ctx, client, key)
		if err != nil {
			return nil, err
		}

		values[i] = value
	}

	return values, nil
}

func (r *RedisV9) getValueFromCluster(ctx context.Context,
	client *redis.ClusterClient,
	key string,
) (interface{}, error) {
	cmd := client.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	val := cmd.Val()
	if val == "" {
		return nil, nil
	}

	return val, nil
}

func (r *RedisV9) getMultiStandalone(ctx context.Context, client *redis.Client, keys []string) ([]interface{}, error) {
	cmd := client.MGet(ctx, keys...)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

// GetKeysAndValuesWithFilter returns all keys and their values for a given pattern
func (r *RedisV9) GetKeysAndValuesWithFilter(ctx context.Context,
	pattern string,
) (map[string]interface{}, error) {
	keys, err := r.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	if len(keys) == 0 {
		return result, nil
	}

	values, err := r.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		result[key] = values[i]
	}

	return result, nil
}

// GetKeysWithOpts performs a paginated scan of keys in a Redis database using the SCAN command.
// It receives a cursor map that contains the cursor position for each Redis node in the cluster.
//
// Parameters:
//
//	ctx:       Execution context.
//	searchStr: Pattern for filtering keys (glob-style patterns allowed).
//	cursor:    Map of Redis node addresses to cursor positions for pagination.
//			   In the first iteration, map must be empty or nil.
//	count:     Approximate number of keys to return per scan.
//
// Returns:
//
//	keys:          Slice of keys matching the searchStr pattern.
//	updatedCursor: Updated cursor map for subsequent pagination.
//	continueScan:  Indicates if more keys are available for scanning (true if any cursor is non-zero).
//	err:           Error, if any occurred during execution.
func (r *RedisV9) GetKeysWithOpts(ctx context.Context,
	searchStr string,
	cursor map[string]uint64,
	count int64,
) ([]string, map[string]uint64, bool, error) {
	var keys []string
	var mutex sync.Mutex
	var continueScan bool

	if cursor == nil {
		cursor = make(map[string]uint64)
	}

	switch client := r.client.(type) {
	case *redis.ClusterClient:
		err := client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			currentCursor, exists := cursor[client.String()]
			if exists && currentCursor == 0 {
				// Cursor is zero, no more keys to scan
				return nil
			}

			localKeys, fkCursor, err := fetchKeysWithCursor(ctx, client, searchStr, cursor[client.String()], count)
			if err != nil {
				return err
			}

			mutex.Lock()
			keys = append(keys, localKeys...)
			cursor[client.String()] = fkCursor
			if fkCursor != 0 {
				continueScan = true
			}
			mutex.Unlock()

			return nil
		})

		if errors.Is(err, redis.ErrClosed) {
			return keys, cursor, continueScan, temperr.ClosedConnection
		}

		if err != nil {
			return keys, cursor, continueScan, err
		}

	case *redis.Client:
		localKeys, fkCursor, err := fetchKeysWithCursor(ctx, client, searchStr, cursor[client.String()], int64(count))
		if err != nil {
			if errors.Is(err, redis.ErrClosed) {
				return localKeys, cursor, continueScan, temperr.ClosedConnection
			}

			return localKeys, cursor, continueScan, err
		}

		cursor[client.String()] = fkCursor

		if fkCursor != 0 {
			continueScan = true
		}
		keys = localKeys

	default:
		return nil, cursor, continueScan, temperr.InvalidRedisClient
	}

	return keys, cursor, continueScan, nil
}

func (r *RedisV9) SetIfNotExist(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, temperr.KeyEmpty
	}

	res := r.client.SetNX(ctx, key, value, expiration)
	if res.Err() != nil {
		return false, res.Err()
	}

	return res.Val(), nil
}

func fetchKeysWithCursor(ctx context.Context,
	client redis.UniversalClient,
	pattern string,
	cursor uint64,
	count int64,
) ([]string, uint64, error) {
	var keys []string

	var err error

	keys, cursor, err = client.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, 0, err
	}

	return keys, cursor, nil
}

func fetchAllKeys(ctx context.Context,
	client redis.UniversalClient,
	pattern string,
) ([]string, error) {
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	return keys, iter.Err()
}
