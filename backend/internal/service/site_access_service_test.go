package service

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// knownHash returns SHA256 hex of s for comparison in tests.
func knownHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func TestHashSiteAnswer_Deterministic(t *testing.T) {
	h1 := hashSiteAnswer("mysecret")
	h2 := hashSiteAnswer("mysecret")
	assert.Equal(t, h1, h2, "same input must always produce same hash")
}

func TestHashSiteAnswer_CaseInsensitive(t *testing.T) {
	lower := hashSiteAnswer("answer")
	upper := hashSiteAnswer("ANSWER")
	mixed := hashSiteAnswer("AnSwEr")
	assert.Equal(t, lower, upper, "uppercase must match lowercase")
	assert.Equal(t, lower, mixed, "mixed case must match lowercase")
}

func TestHashSiteAnswer_WhitespaceTrimmed(t *testing.T) {
	plain := hashSiteAnswer("answer")
	padded := hashSiteAnswer("  answer  ")
	assert.Equal(t, plain, padded, "leading/trailing spaces must be stripped before hashing")
}

func TestHashSiteAnswer_DifferentInputsProduceDifferentHashes(t *testing.T) {
	h1 := hashSiteAnswer("answer1")
	h2 := hashSiteAnswer("answer2")
	assert.NotEqual(t, h1, h2, "different answers must not collide")
}

func TestHashSiteAnswer_KnownValue(t *testing.T) {
	// "hello" lowercased + trimmed is still "hello"
	expected := knownHash("hello")
	assert.Equal(t, expected, hashSiteAnswer("hello"))
	assert.Equal(t, expected, hashSiteAnswer("HELLO"))
	assert.Equal(t, expected, hashSiteAnswer("  Hello  "))
}

func TestHashSiteAnswer_EmptyString(t *testing.T) {
	h := hashSiteAnswer("")
	expected := knownHash("")
	assert.Equal(t, expected, h, "empty string should hash to known SHA256 of empty string")
}
