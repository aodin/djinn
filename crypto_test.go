package djinn

import (
	"bytes"
	"testing"
)

func TestSaltedHMAC(t *testing.T) {
	salt := []byte(`django.contrib.sessionsSessionStore`)
	secret := []byte(`xsy!9deorcwbk!&=u33!ixik-r9c1@sf6tz0jnb*ce9ipe)e&m`)
	data := []byte(`{"_auth_user_backend":"django.contrib.auth.backends.ModelBackend","_auth_user_id":1}`)
	hmacd := SaltedHMAC(salt, secret, data)
	expected := []byte(`1e283abb5af349be094771ed87c1529dd5a4de65`)
	if bytes.Compare(hmacd, expected) != 0 {
		t.Errorf("Unexpected salted hmac: %s != %s", hmacd, expected)
	}
}

func TestGetRandomString(t *testing.T) {
	// Generate a random string, ignore the results
	// TODO Test the entropy?
	GetRandomString(32)
}

func BenchmarkGetRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRandomString(32)
	}
}
