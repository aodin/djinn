package djinn

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"
)

// Calculate the HMAC using the salt and secret as the key.
// The returned byte array will be hex encoded
func SaltedHMAC(salt, secret, data []byte) []byte {
	// Calculate the SHA1 digest of the secret + salt
	h := sha1.New()
	h.Write(salt)
	h.Write(secret)
	key := h.Sum(nil)

	// Create the HMAC
	hmacd := hmac.New(sha1.New, key)
	hmacd.Write(data)
	b := hmacd.Sum(nil)

	// Encode as hex
	// Destination array must have its length specified or it will panic:
	// panic: runtime error: index out of range [recovered]
	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)
	return dst
}

func EncodeBase64Bytes(input []byte) []byte {
	var buf bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &buf)
	e.Write(input)
	e.Close()
	return buf.Bytes()
}

// TODO errors instead of panic?
func RandomBytes(length int) []byte {
	salt := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		panic("djinn: could not generate random bytes")
	}
	return salt
}

// TODO standardize the usage of []byte arrays versus string
func ConstantTimeStringCompare(v1, v2 string) bool {
	// Reimplementation of crypto.subtle.ConstantTimeCompare
	b1 := []byte(v1)
	b2 := []byte(v2)
	if len(b1) != len(b2) {
		return false
	}
	var result byte
	for i := 0; i < len(b1); i++ {
		result |= b1[i] ^ b2[i]
	}
	return subtle.ConstantTimeByteEq(result, 0) == 1
}

func EncodeBase64String(input []byte) string {
	var buf bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &buf)
	e.Write(input)
	e.Close()
	return buf.String()
}

var validChars = []byte(`abcdefghijklmnopqrstuvwxyz0123456789`)

func GetRandomString(n int) string {
	// TODO Different valid characters
	validLen := big.NewInt(int64(len(validChars)))
	dst := make([]byte, n)

	for i, _ := range dst {
		var index int64
		for {
			r, err := rand.Int(rand.Reader, validLen)
			if err == nil {
				index = r.Int64()
				break
			}
		}
		dst[i] = validChars[index]
	}
	return string(dst)
}
