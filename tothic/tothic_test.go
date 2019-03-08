package tothic

import (
	"os"
	"reflect"
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
