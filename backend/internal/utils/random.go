package utils

import (
	"crypto/rand"
	"strings"
)

const digits = "0123456789"

// RandomDigits generates a cryptographically secure random string
// consisting of digits (0–9) of the given length.
func RandomDigits(length int) string {
	if length <= 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(length)

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	for _, v := range b {
		sb.WriteByte(digits[int(v)%len(digits)])
	}

	return sb.String()
}
