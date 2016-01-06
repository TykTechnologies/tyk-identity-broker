/* package backends provides different storage back ends for the configuration of a
TAP node. Backends ned only be k/v stores. The in-memory provider is given as an example and usefule for testing
*/
package backends

import (
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

// InMemoryBackend implements tap.AuthRegisterBackend to store profile configs in memory
type InMemoryBackend struct {
	kv map[string]interface{}
}

// Init will create the initial in-memory store structures
func (m *InMemoryBackend) Init(config interface{}) {
	log.Info("[IN-MEMORY STORE] Initialised")
	m.kv = make(map[string]interface{})
}

// SetKey will set the value of a key in the map
func (m *InMemoryBackend) SetKey(key string, val interface{}) error {
	if m.kv == nil {
		return errors.New("Store not initialised!")
	}

	asByte, encErr := json.Marshal(val)
	if encErr != nil {
		return encErr
	}

	m.kv[key] = asByte
	return nil
}

// SetKey will set the value of a key in the map
func (m *InMemoryBackend) DeleteKey(key string) error {
	delete(m.kv, key)
	return nil
}

// GetKey will retuyrn the value of a key as an interface
func (m *InMemoryBackend) GetKey(key string, target interface{}) error {
	v, ok := m.kv[key]

	if !ok {
		return errors.New("Not found")
	}

	decErr := json.Unmarshal(v.([]byte), target)
	if decErr != nil {
		return decErr
	}

	return nil
}
