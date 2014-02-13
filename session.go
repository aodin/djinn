package djinn

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	SessionDoesNotExist = errors.New("djinn: session does not exist")
	MultipleSessions    = errors.New("djinn: multiple sessions were returned")
	BadSessionData      = errors.New("djinn: improperly formatted session data")
	InvalidHMAC         = errors.New("djinn: the session data hmac is invalid")
)

// django_session
type Session struct {
	Key     string    `db:"session_key"`
	Data    string    `db:"session_data"`
	Expires time.Time `db:"expire_date"`
	manager *SessionManager
}

func (s *Session) String() string {
	return s.Key
}

func (s *Session) Delete() error {
	// TODO There must be a non-nil manager and database connection
	query := fmt.Sprintf(
		`DELETE FROM "%s" WHERE "%s" = %s`,
		s.manager.table,
		s.manager.primary,
		s.manager.db.dialect.Parameter(0),
	)
	_, err := s.manager.db.Exec(query, s.Key)
	return err
}

type SessionManager struct {
	*Manager
}

// The global session manager
// Build columns and primary keys dynamically - on init?
var Sessions = &SessionManager{
	&Manager{
		db:      &connection,
		table:   "django_session",
		columns: []string{"session_key", "session_data", "expire_date"},
		primary: "session_key",
	},
}

// Get a session with an exact matching key and expire date greater than now
func (m *SessionManager) Get(key string) (*Session, error) {
	now := time.Now()

	query := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE "session_key" = %s AND "expire_date" >= %s`,
		m.db.JoinColumns(m.columns),
		m.table,
		m.db.dialect.Parameter(0),
		m.db.dialect.Parameter(1),
	)

	// Don't bother with a destination interface
	rows, err := m.db.Query(query, key, now)
	if err != nil {
		return nil, err
	}

	// One and only one session should be returned
	if !rows.Next() {
		return nil, SessionDoesNotExist
	}
	s := &Session{
		manager: m,
	}
	if err := rows.Scan(&s.Key, &s.Data, &s.Expires); err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, MultipleSessions
	}
	return s, nil
}

// Determine if a session with the given key exists in the database
func (m *SessionManager) Exists(key string) (exists bool, err error) {
	query := fmt.Sprintf(
		`SELECT EXISTS(SELECT 1 FROM "%s" WHERE "session_key" = %s LIMIT 1)`,
		m.table,
		m.db.dialect.Parameter(0),
	)
	err = m.db.QueryRow(query, key).Scan(&exists)
	return
}

func (m *SessionManager) Create(userId int64) (*Session, error) {
	// Create the session data
	data := &SessionData{
		AuthUserBackend: "django.contrib.auth.backends.ModelBackend",
		AuthUserId:      userId,
	}

	// Encode the session data using the configuration salt and secret
	encoded, err := data.Encode(
		[]byte(config.SessionSalt),
		[]byte(config.Secret),
	)
	if err != nil {
		return nil, err
	}

	// Generate a random key - worst case is O(infinity)!
	// But with 36 ** 32 possibilities, we'll need 10 septillion sessions
	// before we hit the birthday bound
	var key string
	for {
		key = GetRandomString(32)
		// Confirm that this key does not already exist
		exists, err := m.Exists(key)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
	}

	// Build the Session
	session := &Session{
		Key:     key,
		Data:    string(encoded),
		Expires: time.Now().Add(config.SessionCookieAge),
		manager: m,
	}

	// Create the query
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		m.table,
		m.db.JoinColumns(m.columns),
		m.db.BuildParameters(m.columns),
	)
	_, err = m.db.Exec(query, &session.Key, &session.Data, &session.Expires)
	// Return nil on error - don't return a session if it wasn't created
	if err != nil {
		return nil, err
	}
	return session, nil
}

type SessionData struct {
	AuthUserBackend string `json:"_auth_user_backend"`
	AuthUserId      int64  `json:"_auth_user_id"`
}

// TODO Encode to bytes?
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

func DecodeSessionData(salt, secret []byte, encoded string) (*SessionData, error) {
	// Decode the base64 data
	// If you try to keep it as byte arrays, the DecodedLen method will
	// return a maximum and there may be additional zero bytes
	data, err := base64.StdEncoding.DecodeString(encoded)
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
	// As of Django 1.6, the default session serializer is JSON
	var sessionData SessionData
	if err = json.Unmarshal(parts[1], &sessionData); err != nil {
		return nil, BadSessionData
	}

	return &sessionData, nil
}
