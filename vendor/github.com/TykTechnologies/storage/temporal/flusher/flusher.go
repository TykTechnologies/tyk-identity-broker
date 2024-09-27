package flusher

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

type Flusher = model.Flusher

func NewFlusher(conn model.Connector) (Flusher, error) {
	switch conn.Type() {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithConnection(conn)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
