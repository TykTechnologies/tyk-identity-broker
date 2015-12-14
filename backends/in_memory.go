package backends

import (
	"encoding/json"
	"errors"
)

type InMemoryBackend struct {
	kv map[string]interface{}
}

func (m *InMemoryBackend) Init() {
	m.kv = make(map[string]interface{})
}

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
