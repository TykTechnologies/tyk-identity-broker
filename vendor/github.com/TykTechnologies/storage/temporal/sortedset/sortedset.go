package sortedset

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

type SortedSet = model.SortedSet

var _ SortedSet = (*redisv9.RedisV9)(nil)

func NewSortedSet(conn model.Connector) (SortedSet, error) {
	switch conn.Type() {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithConnection(conn)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
