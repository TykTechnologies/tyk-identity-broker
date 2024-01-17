package initializer

import (
	"testing"

	temporal "github.com/TykTechnologies/storage/temporal/keyvalue"
	"github.com/TykTechnologies/tyk-identity-broker/backends"
	"github.com/stretchr/testify/assert"
)

func TestCreateBackendFromRedisConn(t *testing.T) {
	var kv temporal.KeyValue
	keyPrefix := "test-prefix"

	// Call the function
	result := CreateBackendFromRedisConn(kv, keyPrefix)

	// Assert that result is not nil
	assert.NotNil(t, result)
	redisBackend, ok := result.(*backends.RedisBackend)
	assert.True(t, ok)

	// Assert that the KeyPrefix is correctly set
	assert.Equal(t, keyPrefix, redisBackend.KeyPrefix)
}
