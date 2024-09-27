package redisv9

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// AddScoredMember adds a member with a specific score to a sorted set in Redis.
// It returns the number of elements added to the sorted set, which is either 0 or 1.
func (r *RedisV9) AddScoredMember(ctx context.Context, key, member string, score float64) (int64, error) {
	return r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Result()
}

// GetMembersByScoreRange retrieves members and their scores from a Redis sorted set
// within the given score range specified by min and max.
// It returns slices of members and their scores, and an error if any occurs during retrieval.
func (r *RedisV9) GetMembersByScoreRange(ctx context.Context, key, min, max string) ([]interface{}, []float64, error) {
	results, err := r.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
	if err != nil {
		return nil, nil, err
	}

	members := make([]interface{}, len(results))
	scores := make([]float64, len(results))

	for i, z := range results {
		members[i] = z.Member
		scores[i] = z.Score
	}

	return members, scores, nil
}

// RemoveMembersByScoreRange removes members from a Redis sorted set within a specified score range.
// It returns the number of members removed from the sorted set.
func (r *RedisV9) RemoveMembersByScoreRange(ctx context.Context, key, min, max string) (int64, error) {
	return r.client.ZRemRangeByScore(ctx, key, min, max).Result()
}
