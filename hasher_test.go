package djinn

import (
	"crypto/sha512"
	"testing"
)

func ExpectPanic(t *testing.T, g func(a string, h Hasher), a string, h Hasher) {
	defer func() {
		if x := recover(); x == nil {
			t.Error("Expected a panic, but one did not occur")
		}
	}()
	g(a, h)
}

func TestHasherRegistry(t *testing.T) {
	// The pbkdf2 hashers should be registered
	_, err := GetHasher("pbkdf2_sha256")
	if err != nil {
		t.Error(err)
	}
	_, err = GetHasher("pbkdf2_sha256")
	if err != nil {
		t.Error(err)
	}

	// New hashers should be able to be registered
	pbkdf2_sha512 := NewPBKDF2Hasher("pbkdf2_sha512", 12000, sha512.New)
	// It will panic if registration fails
	RegisterHasher(pbkdf2_sha512.algorithm, pbkdf2_sha512)

	// No duplicates!
	ExpectPanic(t, RegisterHasher, pbkdf2_sha512.algorithm, pbkdf2_sha512)

	// Nor nil!
	ExpectPanic(t, RegisterHasher, "nil", nil)
}
