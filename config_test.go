package djinn

import (
	"testing"
)

// time.Duration fields must be in number of nanoseconds
var exampleConfig = []byte(`{
	"SECRET_KEY": "I AM A SECRET KEY",
	"SESSION_COOKIE_AGE": 21600000000000,
	"SESSION_COOKIE_SECURE": true
}`)

func TestConfig(t *testing.T) {
	c, err := ParseConfig(exampleConfig)
	if err != nil {
		t.Fatal(err)
	}
	expectString(t, c.Secret, "I AM A SECRET KEY")
	expectDuration(t, c.SessionCookieAge, "6h")
	if !c.SessionCookieSecure {
		t.Error("Session Cookie Secure was unexpectedly set false")
	}
}
