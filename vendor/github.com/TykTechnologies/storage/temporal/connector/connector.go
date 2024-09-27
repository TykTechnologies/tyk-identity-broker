package connector

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

var WithRedisConfig = model.WithRedisConfig

var _ model.Connector = (*redisv9.RedisV9)(nil)

// NewConnector returns a new connector based on the type. You have to specify the connector Configuration as an Option.
func NewConnector(connType string, options ...model.Option) (model.Connector, error) {
	switch connType {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithOpts(options...)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
