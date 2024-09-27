package list

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

type List = model.List

var _ List = (*redisv9.RedisV9)(nil)

func NewList(conn model.Connector) (List, error) {
	switch conn.Type() {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithConnection(conn)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
