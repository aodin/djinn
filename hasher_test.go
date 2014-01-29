package djinn

import (
	"crypto/sha512"
	"testing"
)

func expectPanic(t *testing.T, g func(a string, h Hasher), a string, h Hasher) {
	defer func() {
		if x := recover(); x == nil {
			t.Error("Expected a panic, but one did not occur")
		}
	}()
	g(a, h)
}

func TestMD5Hasher(t *testing.T) {
	md5, err := GetHasher("md5")
	if err != nil {
		t.Fatal(err)
	}

	hash := md5.Encode("admin", "BD03RxMbKE9o")
	expectString(t, hash, "md5$BD03RxMbKE9o$7de5de2fb33be2b11dee2e016517df5a")

	if !md5.Verify("admin", hash) {
		t.Error("Could not verify MD5 password hash")
	}
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
	expectPanic(t, RegisterHasher, pbkdf2_sha512.algorithm, pbkdf2_sha512)

	// Nor nil!
	expectPanic(t, RegisterHasher, "nil", nil)
}
