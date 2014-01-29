package djinn

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

type Hasher interface {
	Encode(string, string) string
	Salt() string
	Verify(string, string) bool
	Algorithm() string
}

func MakePassword(h Hasher, cleartext string) string {
	return h.Encode(cleartext, h.Salt())
}

func CheckPassword(h Hasher, cleartext, encoded string) bool {
	return h.Verify(cleartext, encoded)
}

var hashers = make(map[string]Hasher)

func RegisterHasher(name string, hasher Hasher) {
	if hasher == nil {
		panic("djinn: attempting to register a nil Hasher")
	}
	if _, duplicate := hashers[name]; duplicate {
		panic("djinn: RegisterHasher called twice for Hasher " + name)
	}
	hashers[name] = hasher
}

func GetHasher(name string) (Hasher, error) {
	hasher, ok := hashers[name]
	if !ok {
		return nil, fmt.Errorf("djinn: unknown hasher %s (did you remember to import it?)", name)
	}
	return hasher, nil
}

// The BaseHasher struct is the parent of all included Hashers
type BaseHasher struct {
	algorithm string
}

// Create a random string
func (b *BaseHasher) Salt() string {
	return base64.StdEncoding.EncodeToString(RandomBytes(9))
}

func (b *BaseHasher) Algorithm() string {
	return b.algorithm
}

// Create an MD5 hash that should never be used except for testing
type MD5Hasher struct {
	BaseHasher
}

func (m *MD5Hasher) Encode(cleartext, salt string) string {
	h := md5.New()
	h.Write([]byte(salt))
	h.Write([]byte(cleartext))
	b := h.Sum(nil)

	// Encode as hex
	return strings.Join([]string{m.algorithm, salt, hex.EncodeToString(b)}, "$")
}

func (m *MD5Hasher) Verify(cleartext, encoded string) bool {
	// Split the saved hash apart
	parts := strings.SplitN(encoded, "$", 3)

	// TODO Errors? What about improperly formatted hashes?
	if len(parts) != 3 {
		return false
	}
	if parts[0] != m.algorithm {
		return false
	}

	// Re-create the hash using the cleartext and salt (parts[1])
	// TODO There is duplicate with the Encode method
	h := md5.New()
	h.Write([]byte(parts[1]))
	h.Write([]byte(cleartext))
	b := h.Sum(nil)
	rehash := hex.EncodeToString(b)

	// Perform a constant time comparison between the new and old hashes
	return ConstantTimeStringCompare(rehash, parts[2])
}

func NewMD5Hasher() *MD5Hasher {
	return &MD5Hasher{BaseHasher{algorithm: "md5"}}
}

func init() {
	md5 := NewMD5Hasher()
	RegisterHasher(md5.algorithm, md5)
}
