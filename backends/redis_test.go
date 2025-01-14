package backends

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-identity-broker/tap"

	mocks "github.com/TykTechnologies/storage/temporal/tempmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockRedisBackend(t *testing.T) (*RedisBackend, *mocks.KeyValue) {
	testObj := mocks.NewKeyValue(t)
	rb := &RedisBackend{
		kv:        testObj,
		config:    &RedisConfig{},
		KeyPrefix: "key-prefix",
	}
	return rb, testObj
}

func TestConnect(t *testing.T) {
	testObj := mocks.NewKeyValue(t)

	rb := RedisBackend{
		kv:     testObj,
		config: &RedisConfig{},
	}

	err := rb.Connect()

	assert.Nil(t, err)
	assert.NotNil(t, rb.kv, "key-value store should not be nil")
}

// TestCleanKey tests the cleanKey function
func TestCleanKey(t *testing.T) {
	r, _ := mockRedisBackend(t)
	r.KeyPrefix = "prefix_"
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

// TestFixKey tests the fixKey function
func TestFixKey(t *testing.T) {
	r, _ := mockRedisBackend(t)
	r.KeyPrefix = "prefix_"
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

func TestRedisInit(t *testing.T) {

	testCases := []struct {
		name      string
		config    interface{}
		shouldErr bool
	}{
		{
			name:      "invalid config - numeric",
			config:    1111,
			shouldErr: true,
		},
		{
			name:      "invalid config - random string",
			config:    "some-invalid-config",
			shouldErr: true,
		},
		{
			name: "valid config",
			// change some configs
			config: RedisConfig{
				MaxIdle:               1,
				MaxActive:             0,
				MasterName:            "some-master",
				Database:              1,
				Username:              "testUser",
				Password:              "s3cr3t",
				UseSSL:                true,
				SSLInsecureSkipVerify: true,
				Port:                  5000,
				MaxVersion:            "1.0",
				MinVersion:            "1.0",
			},
			shouldErr: false,
		},
		{
			name: "unvalid TLS MAX/Min Version",
			// change some configs
			config: RedisConfig{
				MaxIdle:               1,
				MaxActive:             0,
				MasterName:            "some-master",
				Database:              1,
				Username:              "testUser",
				Password:              "s3cr3t",
				UseSSL:                true,
				SSLInsecureSkipVerify: true,
				Port:                  5000,
				MaxVersion:            "xxx",
				MinVersion:            "yyy",
			},
			shouldErr: true,
		},
		{
			name: "Min version is greater than max version",
			// change some configs
			config: RedisConfig{
				MaxIdle:               1,
				MaxActive:             0,
				MasterName:            "some-master",
				Database:              1,
				Username:              "testUser",
				Password:              "s3cr3t",
				UseSSL:                true,
				SSLInsecureSkipVerify: true,
				Port:                  5000,
				MaxVersion:            "1.0",
				MinVersion:            "1.2",
			},
			shouldErr: true,
		},
		{
			name:      "invalid config - non marshable",
			config:    make(chan int),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := mockRedisBackend(t)

			err := r.Init(tc.config)

			didErr := err != nil
			assert.Equal(t, tc.shouldErr, didErr)

			if !tc.shouldErr {
				assert.Equal(t, tc.config, *r.config)
			}
		})
	}
}

func TestRedisBackend_SetDb(t *testing.T) {
	// Mock KeyValue instance
	testObj := mocks.NewKeyValue(t)
	testObj.Test(t)

	// Create an instance of RedisBackend
	r := &RedisBackend{
		// Initialize other necessary fields, if any
	}

	// Call SetDb with the mock KeyValue
	r.SetDb(testObj)

	// Assertions
	assert.Equal(t, testObj, r.kv, "KeyValue instance not set correctly in RedisBackend")
}

func TestRedis_SetKey(t *testing.T) {
	rb, testObj := mockRedisBackend(t)

	keyName := "key"
	orgId := "orgId"
	value := "test-val"
	var ttl time.Duration

	testObj.On("Set", mock.Anything, rb.KeyPrefix+keyName, value, ttl).Return(nil)
	err := rb.SetKey(keyName, orgId, value)
	assert.Nil(t, err)
	testObj.AssertExpectations(t)
}

func TestRedis_GetKey(t *testing.T) {
	// Setting up mocks
	rb, testObj := mockRedisBackend(t)

	// Preparing test data
	testProfile := tap.Profile{
		ID:    "some-profile",
		OrgID: "test-org",
	}

	bytes, err := json.Marshal(testProfile)
	assert.Nil(t, err)

	keyName := "key"
	orgId := "orgId"
	value := string(bytes)
	var ttl time.Duration

	// Setting up expectations for the mock object
	testObj.On("Set", mock.Anything, rb.KeyPrefix+keyName, value, ttl).Return(nil)
	testObj.On("Get", mock.Anything, rb.KeyPrefix+keyName).Return(value, nil)

	// Executing the function under test
	err = rb.SetKey(keyName, orgId, value)
	assert.Nil(t, err)

	var newVal tap.Profile
	err = rb.GetKey(keyName, orgId, &newVal)
	assert.Nil(t, err)

	// Verifying that expectations were met
	testObj.AssertExpectations(t)
}

func TestRedis_DeleteKey(t *testing.T) {
	rb, testObj := mockRedisBackend(t)
	key := "keyName"
	orgId := "orgId"

	testObj.On("Delete", mock.Anything, rb.KeyPrefix+key).Return(nil)

	err := rb.DeleteKey(key, orgId)
	assert.Nil(t, err)
	testObj.AssertExpectations(t)
}

func TestRedis_GetAll(t *testing.T) {
	rb, testObj := mockRedisBackend(t)

	orgId := "test-org"

	testObj.On("Keys", mock.Anything, rb.KeyPrefix).Return([]string{}, errors.New("pulling keys"))
	keys := rb.GetAll(orgId)
	assert.Len(t, keys, 0)
}

// Test for toJSONString function
func TestToJSONString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
		err      bool
	}{
		// Test cases with expected outputs
		{input: map[string]string{"key": "value"}, expected: `{"key":"value"}`, err: false},
		{input: []int{1, 2, 3}, expected: `[1,2,3]`, err: false},
		{input: "Hello, world!", expected: `"Hello, world!"`, err: false},
		{input: nil, expected: "null", err: false},
		{input: struct{ Name string }{Name: "Alice"}, expected: `{"Name":"Alice"}`, err: false},

		// Test case expected to fail (circular reference)
		{input: func() {}, expected: "", err: true}, // Function types cannot be marshaled
	}

	for _, test := range tests {
		result, err := toJSONString(test.input)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for input %v, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %v: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("Expected %s for input %v, got %s", test.expected, test.input, result)
			}
		}
	}
}
