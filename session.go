package djinn

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var (
	BadSessionData = errors.New("djinn: improperly formatted session data")
	InvalidHMAC    = errors.New("djinn: the session data hmac is invalid")
)

// django_session
type Session struct {
	Key     string     `db:"session_key"`
	Data    string     `db:"session_data"`
	Expires *time.Time `db:"expire_date"`
}

func (s *Session) String() string {
	return s.Key
}

type SessionData struct {
	AuthUserBackend string `json:"_auth_user_backend"`
	AuthUserId      int64  `json:"_auth_user_id"`
}

func (s *SessionData) Encode(salt, secret []byte) ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Calculate the salted hmac of the json encoded data
	hmacd := SaltedHMAC(salt, secret, data)
	b := bytes.Join([][]byte{hmacd, data}, []byte{':'})

	// Encode as base64
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(dst, b)
	return dst, nil
}

func DecodeSessionData(salt, secret, encoded []byte) (*SessionData, error) {
	// Decode the base64 data
	// If you try to keep it as byte arrays, the DecodedLen method will
	// return a maximum and there may be additional zero bytes
	// data := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	// _, err := base64.StdEncoding.Decode(data, encoded)
	data, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, err
	}

	// Split the data at the colon
	parts := bytes.SplitN(data, []byte{':'}, 2)
	if len(parts) != 2 {
		return nil, BadSessionData
	}

	// Re-calculate the HMAC
	rehmac := SaltedHMAC(salt, secret, parts[1])
	// log.Println(rehmac)

	// Constant time compare the given and calculated hmacs
	if !hmac.Equal(parts[0], rehmac) {
		return nil, InvalidHMAC
	}

	// Decode the session data
	// Python's pickle is close enough to json that default data works
	var sessionData SessionData
	if err = json.Unmarshal(parts[1], &sessionData); err != nil {
		return nil, BadSessionData
	}

	return &sessionData, nil
}
