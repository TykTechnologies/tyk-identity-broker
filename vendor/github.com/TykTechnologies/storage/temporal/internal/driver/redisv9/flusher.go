package redisv9

import (
	"context"

	"github.com/TykTechnologies/storage/temporal/temperr"
	"github.com/redis/go-redis/v9"
)

func (r *RedisV9) FlushAll(ctx context.Context) error {
	switch client := r.client.(type) {
	case *redis.ClusterClient:
		return client.ForEachMaster(ctx, func(context context.Context, client *redis.Client) error {
			return client.FlushAll(ctx).Err()
		})
	case *redis.Client:
		return r.client.FlushAll(ctx).Err()
	default:
		return temperr.InvalidHandlerType
	}
}
