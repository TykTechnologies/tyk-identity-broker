package backends

import (
	mocks "github.com/TykTechnologies/storage/temporal/tempmocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func mockRedisBackend() *RedisBackend {
	return &RedisBackend{KeyPrefix: "prefix_"}
}

func TestConnect(t *testing.T) {
	testObj := mocks.NewKeyValue(t)

	rb := RedisBackend{
		kv:     testObj,
		config: &RedisConfig{},
	}

	tkv, err := rb.Connect()

	assert.Nil(t, err)
	assert.NotNil(t, tkv, "key-value store should not be nil")
}

// TestCleanKey tests the cleanKey function
func TestCleanKey(t *testing.T) {
	r := mockRedisBackend()
	tests := []struct {
		keyName string
		want    string
	}{
		{"prefix_key1", "key1"},
		{"prefix_key2", "key2"},
		{"key3", "key3"},
	}

	for _, tt := range tests {
		t.Run(tt.keyName, func(t *testing.T) {
			if got := r.cleanKey(tt.keyName); got != tt.want {
				t.Errorf("cleanKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHashKey tests the hashKey function
func TestHashKey(t *testing.T) {
	r := mockRedisBackend()
	input := "testKey"
	want := "testKey" // hashKey returns the input as is
	if got := r.hashKey(input); got != want {
		t.Errorf("hashKey() = %v, want %v", got, want)
	}
}

// TestFixKey tests the fixKey function
func TestFixKey(t *testing.T) {
	r := mockRedisBackend()
	tests := []struct {
		keyName string
		want    string
	}{
		{"key1", "prefix_key1"},
		{"key2", "prefix_key2"},
	}

	for _, tt := range tests {
		t.Run(tt.keyName, func(t *testing.T) {
			if got := r.fixKey(tt.keyName); got != tt.want {
				t.Errorf("fixKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
