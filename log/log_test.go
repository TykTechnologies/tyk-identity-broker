package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestGetIsPureGetter verifies that calling Get() multiple times never mutates
// the logger's level. This is a regression test for the previous behaviour
// where Get() read TYK_LOGLEVEL and called SetLevel on every invocation,
// which would silently reset any level configured by the caller.
func TestGetIsPureGetter(t *testing.T) {
	Get().SetLevel(logrus.ErrorLevel)
	Get()

	assert.Equal(t, logrus.ErrorLevel, Get().Level)
}

// TestSetLoggerLevelNotOverwritten verifies that after an embedder injects its
// own logger via SetLogger, subsequent Get() calls do not overwrite that
// logger's level. This covers the dashboard embedding scenario where
// TYK_DB_LOGLEVEL=error was being ignored because TIB's Get() reset the
// shared logger back to InfoLevel.
func TestSetLoggerLevelNotOverwritten(t *testing.T) {
	t.Cleanup(func() {
		SetLogger(logrus.New())
	})

	injected := logrus.New()
	injected.SetLevel(logrus.ErrorLevel)

	SetLogger(injected)

	Get()

	assert.Equal(t, logrus.ErrorLevel, Get().Level)
}
