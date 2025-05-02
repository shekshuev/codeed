package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomDigits(t *testing.T) {
	t.Run("returns string of correct length", func(t *testing.T) {
		code := RandomDigits(6)
		assert.Len(t, code, 6)
	})

	t.Run("returns only digits", func(t *testing.T) {
		code := RandomDigits(10)
		for _, c := range code {
			assert.True(t, c >= '0' && c <= '9')
		}
	})

	t.Run("returns empty string for zero length", func(t *testing.T) {
		code := RandomDigits(0)
		assert.Equal(t, "", code)
	})

	t.Run("returns different values on each call", func(t *testing.T) {
		code1 := RandomDigits(8)
		code2 := RandomDigits(8)
		assert.NotEqual(t, code1, code2)
	})
}
