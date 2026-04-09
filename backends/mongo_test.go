package backends

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInit tests the Init method of the MongoBackend
func TestMongoInit(t *testing.T) {
	// Create an instance of MongoBackend
	m := MongoBackend{}

	// Call the Init method
	err := m.Init(nil)

	// Assert that Init returns nil
	assert.Nil(t, err)
}
