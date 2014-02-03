package djinn

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"time"
)

type Config struct {
	LoginURL              string        `json:"LOGIN_URL"`
	PasswordHasher        string        `json:"PASSWORD_HASHER"`
	Secret                string        `json:"SECRET_KEY"`
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

func (c Config) Merge(merge Config) Config {
	// Overwrite the existing values in "c" with those given in "merge"
	// Any zero files in the "merge" file will be ignored
	// TODO Use reflect to perform these checks
	// TODO What if we actually want the zero initialization?
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

func SetConfig(c Config) {
	config = c
}

// TODO Default values for the config?
func ParseConfig(contents []byte) (c Config, err error) {
	err = json.Unmarshal(contents, &c)
	return
}

func LoadConfig(path string) (c Config, err error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return ParseConfig(contents)
}

// The configuration can be set by command line flags or a JSON configuration
// file. The JSON configuration will take precendence over the default
// configuration and the command line flags over both.

func init() {
	// TODO Parse a JSON config if one was given
	// TODO Merge the parsed configuration with the default

	// Overwrite with any command line options
	// Set the default value for the SECRET KEY to the existing secret
	flag.StringVar(
		&config.Secret,
		"secret",
		config.Secret, // This must be set or bad news bears
		"The SECRET_KEY that will be used to encode session data",
	)
	flag.Parse()
}

// Provide read-only access to the default configuration
func Settings() Config {
	return config
}

// Set the SECRET KEY on the default configuration
func SetSecret(secret string) {
	config.Secret = secret
}
