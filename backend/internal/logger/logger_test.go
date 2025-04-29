package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLogger_ReturnsSingleton(t *testing.T) {
	logger1 := NewLogger()
	logger2 := NewLogger()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger1.Log)
	assert.Same(t, logger1, logger2, "NewLogger should return the same instance (singleton)")
}

func TestLogger_Initialize_InvalidLevel(t *testing.T) {
	l := &Logger{}
	err := l.initialize("invalid-level")
	assert.Error(t, err)
}

func TestLogger_Initialize_ValidLevel(t *testing.T) {
	l := &Logger{}
	err := l.initialize("debug")
	assert.NoError(t, err)
	assert.IsType(t, &zap.Logger{}, l.Log)
}
