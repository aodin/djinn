package djinn

import (
	"bytes"
	"testing"
)

func TestSessionData_Encode(t *testing.T) {
	salt := []byte(`django.contrib.sessionsSessionStore`)
	secret := []byte(`xsy!9deorcwbk!&=u33!ixik-r9c1@sf6tz0jnb*ce9ipe)e&m`)

	data := &SessionData{
		AuthUserBackend: "django.contrib.auth.backends.ModelBackend",
		AuthUserId:      1,
	}
	d, err := data.Encode(salt, secret)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte(`MWUyODNhYmI1YWYzNDliZTA5NDc3MWVkODdjMTUyOWRkNWE0ZGU2NTp7Il9hdXRoX3VzZXJfYmFja2VuZCI6ImRqYW5nby5jb250cmliLmF1dGguYmFja2VuZHMuTW9kZWxCYWNrZW5kIiwiX2F1dGhfdXNlcl9pZCI6MX0=`)
	if bytes.Compare(d, expected) != 0 {
		t.Errorf("Unexpected encoded session data: %s != %s", d, expected)
	}
}

func TestDecodeSessionData(t *testing.T) {
	salt := []byte(`django.contrib.sessionsSessionStore`)
	secret := []byte(`xsy!9deorcwbk!&=u33!ixik-r9c1@sf6tz0jnb*ce9ipe)e&m`)
	encoded := []byte(`MWUyODNhYmI1YWYzNDliZTA5NDc3MWVkODdjMTUyOWRkNWE0ZGU2NTp7Il9hdXRoX3VzZXJfYmFja2VuZCI6ImRqYW5nby5jb250cmliLmF1dGguYmFja2VuZHMuTW9kZWxCYWNrZW5kIiwiX2F1dGhfdXNlcl9pZCI6MX0=`)

	data, err := DecodeSessionData(salt, secret, encoded)
	if err != nil {
		t.Fatal(err)
	}

	expectedAuthUserBacked := `django.contrib.auth.backends.ModelBackend`
	if data.AuthUserBackend != expectedAuthUserBacked {
		t.Errorf("Unexpected auth user backend: %s != %s", data.AuthUserBackend, expectedAuthUserBacked)
	}

	var expectedAuthUserId int64 = 1
	if data.AuthUserId != expectedAuthUserId {
		t.Errorf("Unexpected auth user id: %d != %d", data.AuthUserId, expectedAuthUserId)
	}
}
