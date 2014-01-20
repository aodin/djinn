package djinn

import (
	"fmt"
)

type Hasher interface {
	Encode(string, string) string
	Salt() string
	Verify(string, string) bool
	Algorithm() string
}

// The default password hasher
// TODO This should be set dynamically
var defaultHasher Hasher

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

func (bH *BaseHasher) Salt() string {
	// Create a random string
	return EncodeBase64String(RandomBytes(9))
}

func (bH *BaseHasher) Algorithm() string {
	return bH.algorithm
}

func NewBaseHasher(algorithm string) BaseHasher {
	return BaseHasher{algorithm: algorithm}
}
