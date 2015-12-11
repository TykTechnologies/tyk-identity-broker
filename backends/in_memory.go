package backends

type InMemoryBackend struct {
	kv map[string]interface{}
}

func (m *InMemoryBackend) Init() {
	m.kv = make(map[string]interface{})
}

func (m *InMemoryBackend) SetKey(key string, val interface{}) error {
	if m.kv == nil {
		return errors.New("Store inot initialised!")
	}

	m.kv[key] = val
	return nil
}
func (m *InMemoryBackend) GetKey(key string) (interface{}, error) {
	v, ok := m.kv[key]

	if !ok {
		return nil, errors.New("Not found")
	}

	return v, nil
}
