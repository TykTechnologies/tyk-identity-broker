package redisv9

import (
	"context"

	"github.com/TykTechnologies/storage/temporal/temperr"
)

// Returns all the members of the set value stored at key
func (r *RedisV9) Members(ctx context.Context, key string) ([]string, error) {
	if key == "" {
		return []string{}, temperr.KeyEmpty
	}

	return r.client.SMembers(ctx, key).Result()
}

// Add the specified members to the set stored at key.
// Specified members that are already a member of this set are ignored.
// If key does not exist, a new set is created before adding the specified members.
// It errors if the key is not a set.
func (r *RedisV9) AddMember(ctx context.Context, key, member string) error {
	if key == "" {
		return temperr.KeyEmpty
	}

	return r.client.SAdd(ctx, key, member).Err()
}

// Remove the specified members from the set stored at key.
// Specified members that are not a member of this set are ignored.
// It errors if the key is not a set.
func (r *RedisV9) RemoveMember(ctx context.Context, key, member string) error {
	if key == "" {
		return temperr.KeyEmpty
	}

	return r.client.SRem(ctx, key, member).Err()
}

// Returns if member is a member of the set stored at key.
func (r *RedisV9) IsMember(ctx context.Context, key, member string) (bool, error) {
	if key == "" {
		return false, temperr.KeyEmpty
	}

	return r.client.SIsMember(ctx, key, member).Result()
}
