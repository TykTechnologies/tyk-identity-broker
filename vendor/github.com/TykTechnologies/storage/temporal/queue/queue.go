package queue

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

type Queue = model.Queue

var _ Queue = (*redisv9.RedisV9)(nil)

func NewQueue(conn model.Connector) (Queue, error) {
	switch conn.Type() {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithConnection(conn)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
