package backends

import "testing"

func TestInMemoryBackend_GetAndSetKey(t *testing.T) {
	backend := &InMemoryBackend{}

	var config interface{}

	backend.Init(config)

	type aStruct struct {
		Thing string
	}

	saveVal := aStruct{Thing: "Test"}
	keyName := "test-key"

	sErr := backend.SetKey(keyName, saveVal)
	if sErr != nil {
		t.Error("Error raised on set key: ", sErr)
	}

	target := aStruct{}
	vErr := backend.GetKey(keyName, &target)

	if vErr != nil {
		t.Error("Error raised on get key: ", vErr)
	}

	if target.Thing != saveVal.Thing {
		t.Error("Expected 'Test' as key val, got: ", target.Thing)
	}
}
