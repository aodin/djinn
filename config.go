package djinn

import (
	"flag"
)

type Config struct {
	PasswordHasher    string `json:"PASSWORD_HASHER"` // default, not full list
	Secret            string `json:"SECRET"`
	SessionSalt       string `json:"SESSION_SALT"`
	SessionCookieName string `json:"SESSION_COOKIE_NAME"`
	// TODO Database configuration(s)
}

var config = Config{
	Secret:            "", // This must be set or bad news bears
	SessionSalt:       "django.contrib.sessionsSessionStore",
	SessionCookieName: "sessionid",
	PasswordHasher:    "pbkdf2_sha256",
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
