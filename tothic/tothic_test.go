package tothic

import (
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestKeyFromEnv(t *testing.T) {
	expected := "SECRET"

	t.Run("with new variable", func(t *testing.T) {
		_ = os.Setenv("TYK_IB_SESSION_SECRET", expected)
		_ = os.Setenv("SESSION_SECRET", "")
		assert(t, expected, KeyFromEnv())
	})

	t.Run("with deprecated", func(t *testing.T) {
		_ = os.Setenv("TYK_IB_SESSION_SECRET", "")
		_ = os.Setenv("SESSION_SECRET", expected)
		assert(t, expected, KeyFromEnv())
	})

	t.Run("with both", func(t *testing.T) {
		_ = os.Setenv("SESSION_SECRET", "SOME_OTHER_SECRET")
		_ = os.Setenv("TYK_IB_SESSION_SECRET", expected)
		assert(t, expected, KeyFromEnv())
	})
}

func assert(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, actual %v", expected, actual)
	}
}

func TestGetState(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert(t, "state", GetState(req))

	req, _ = http.NewRequest(http.MethodGet, "http://localhost?state=FooBar", nil)
	assert(t, "FooBar", GetState(req))

	req, _ = http.NewRequest(http.MethodPost, "http://localhost", nil)
	assert(t, "state", GetState(req))

	req, _ = http.NewRequest(http.MethodPost, "http://localhost?state=FooBar", nil)
	assert(t, "FooBar", GetState(req))

	data := url.Values{}
	data.Add("state", "BarBaz")

	requestBody := data.Encode()

	req, _ = http.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assert(t, "BarBaz", GetState(req))

	req, _ = http.NewRequest(http.MethodPost, "http://localhost?state=FooBar", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assert(t, "FooBar", GetState(req))
}
