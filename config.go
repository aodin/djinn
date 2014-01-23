package djinn

import (
	"flag"
	"time"
)

type Config struct {
	PasswordHasher    string        `json:"PASSWORD_HASHER"`
	Secret            string        `json:"SECRET"`
	SessionSalt       string        `json:"SESSION_SALT"`
	SessionCookieAge  time.Duration `json:"SESSION_COOKIE_AGE"`
	SessionCookieName string        `json:"SESSION_COOKIE_NAME"`
	// TODO Database configuration(s)
	// TODO Specify multiple password hashing algorithms
}

func (c Config) Copy() Config {
	return c
}

var config = Config{
	PasswordHasher:    "pbkdf2_sha256",
	Secret:            "", // This must be set or bad news bears
	SessionSalt:       "django.contrib.sessionsSessionStore",
	SessionCookieAge:  14 * 24 * time.Hour, // 2 weeks
	SessionCookieName: "sessionid",
}

// The configuration can be set by command line flags or a JSON configuration
// file. The JSON configuration will take precendence over the default
// configuration and the command line flags over both.

func init() {
	flag.StringVar(&config.Secret, "secret", "", "The SECRET_KEY that will be used to encode session data")
	flag.Parse()
}

func Settings() Config {
	return config
}
