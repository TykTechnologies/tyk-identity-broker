package temporal

import (
	"github.com/TykTechnologies/storage/temporal/internal/driver/redisv9"
	"github.com/TykTechnologies/storage/temporal/model"
	"github.com/TykTechnologies/storage/temporal/temperr"
)

type KeyValue = model.KeyValue

var _ KeyValue = (*redisv9.RedisV9)(nil)

// NewKeyValue returns a new model.KeyValue storage based on the type of the connector.
func NewKeyValue(conn model.Connector) (KeyValue, error) {
	switch conn.Type() {
	case model.RedisV9Type:
		return redisv9.NewRedisV9WithConnection(conn)
	default:
		return nil, temperr.InvalidHandlerType
	}
}
