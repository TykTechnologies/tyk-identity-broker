/* package backends provides different storage back ends for the configuration of a
TAP node. Backends need to only be k/v stores. The in-memory provider is given as an example and usefule for testing
*/
package backends

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/TykTechnologies/tyk-identity-broker/log"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
	"github.com/sirupsen/logrus"
)

var logger = log.Get()
var inMemoryLogTag = "TIB IN-MEMORY STORE"
var inMemoryLogger = logger.WithField("prefix", inMemoryLogTag)

// InMemoryBackend implements tap.AuthRegisterBackend to store profile configs in memory
type InMemoryBackend struct {
	kv   map[string]interface{}
	lock sync.RWMutex
}

// Init will create the initial in-memory store structures
func (m *InMemoryBackend) Init(config interface{}) error {
	logger = log.Get()
	inMemoryLogger = &logrus.Entry{Logger: logger}
	inMemoryLogger = inMemoryLogger.Logger.WithField("prefix", inMemoryLogTag)

	inMemoryLogger.Info("Initialised")
	m.kv = make(map[string]interface{})
	m.lock = sync.RWMutex{}
	return nil
}

// SetKey will set the value of a key in the map
func (m *InMemoryBackend) SetKey(key string, orgId string, val interface{}) error {
	if m.kv == nil {
		return errors.New("store not initialised!")
	}

	asByte, encErr := json.Marshal(val)
	if encErr != nil {
		return encErr
	}

	m.lock.Lock()
	m.kv[key] = asByte
	m.lock.Unlock()
	return nil
}

// SetKey will set the value of a key in the map
func (m *InMemoryBackend) DeleteKey(key, orgId string) error {
	m.lock.Lock()
	delete(m.kv, key)
	m.lock.Unlock()
	return nil
}

// GetKey will retuyrn the value of a key as an interface
func (m *InMemoryBackend) GetKey(key string, orgId string, target interface{}) error {
	m.lock.RLock()
	defer m.lock.RUnlock()

	v, ok := m.kv[key]

	if !ok {
		return errors.New("not found")
	}

	decErr := json.Unmarshal(v.([]byte), target)
	if decErr != nil {
		return decErr
	}

	return nil
}

func (m *InMemoryBackend) GetAll(orgId string) []interface{} {

	m.lock.RLock()
	defer m.lock.RUnlock()

	target := make([]interface{}, 0)
	for _, v := range m.kv {

		var thisVal tap.Profile
		decErr := json.Unmarshal(v.([]byte), &thisVal)
		if decErr != nil {
			inMemoryLogger.Error(decErr)
		} else {
			target = append(target, thisVal)
		}
	}
	return target
}
