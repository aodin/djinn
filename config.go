package djinn

import (
	"flag"
	"time"
)

type Config struct {
	LoginURL              string        `json:"LOGIN_URL"`
	PasswordHasher        string        `json:"PASSWORD_HASHER"`
	Secret                string        `json:"SECRET"`
	SessionSalt           string        `json:"SESSION_SALT"`
	SessionCookieAge      time.Duration `json:"SESSION_COOKIE_AGE"`
	SessionCookieDomain   string        `json:"SESSION_COOKIE_DOMAIN"`
	SessionCookieHttpOnly bool          `json:"SESSION_COOKIE_HTTPONLY"`
	SessionCookieName     string        `json:"SESSION_COOKIE_NAME"`
	SessionCookiePath     string        `json:"SESSION_COOKIE_PATH"`
	SessionCookieSecure   bool          `json:"SESSION_COOKIE_SECURE"`
	// TODO Database configuration(s)
	// TODO Specify multiple password hashing algorithms
}

func (c Config) Copy() Config {
	return c
}

var config = Config{
	LoginURL:              "/login",
	PasswordHasher:        "pbkdf2_sha256",
	Secret:                "",
	SessionSalt:           "django.contrib.sessionsSessionStore",
	SessionCookieAge:      14 * 24 * time.Hour, // 2 weeks
	SessionCookieDomain:   "",
	SessionCookieHttpOnly: true,
	SessionCookieName:     "sessionid",
	SessionCookiePath:     "/",
	SessionCookieSecure:   false,
}

// The configuration can be set by command line flags or a JSON configuration
// file. The JSON configuration will take precendence over the default
// configuration and the command line flags over both.

func init() {
	flag.StringVar(
		&config.Secret,
		"secret",
		"", // This must be set or bad news bears
		"The SECRET_KEY that will be used to encode session data",
	)
	flag.Parse()
}

func Settings() Config {
	return config
}

func SetSecret(secret string) {
	config.Secret = secret
}
