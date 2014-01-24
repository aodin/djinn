package djinn

import (
	"bytes"
	"crypto/hmac"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	SessionDoesNotExist = errors.New("djinn: session does not exist")
	MultipleSessions    = errors.New("djinn: multiple sessions were returned")
	BadSessionData      = errors.New("djinn: improperly formatted session data")
	InvalidHMAC         = errors.New("djinn: the session data hmac is invalid")
	KeylessSession      = errors.New("djinn: a session must have a key to delete")
)

// django_session
type Session struct {
	Key     string    `db:"session_key"`
	Data    string    `db:"session_data"`
	Expires time.Time `db:"expire_date"`
}

func (s *Session) String() string {
	return s.Key
}

func (s *Session) Delete() error {
	if s.Key == "" {
		return KeylessSession
	}
	// TODO Include a manager object in each session instance
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, Sessions.table)
	_, err := Sessions.db.Exec(query, s.Key)
	return err
}

// The session manager instance that will be populated on init()
var Sessions *SessionManager

type SessionManager struct {
	db      *sql.DB
	table   string
	columns []string
}

// Get a session with an exact matching key and expire date greater than now
func (m *SessionManager) Get(key string) (*Session, error) {
	now := time.Now()

	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE session_key = $1 AND expire_date >= $2`,
		strings.Join(m.columns, ", "),
		m.table,
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
	session := &Session{}
	if err := rows.Scan(&session.Key, &session.Data, &session.Expires); err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, MultipleSessions
	}
	return session, nil
}

// Determine if a session with the given key exists in the database
func (m *SessionManager) Exists(key string) (exists bool, err error) {
	query := fmt.Sprintf(
		`SELECT EXISTS(SELECT 1 FROM %s WHERE session_key = $1 LIMIT 1)`,
		m.table,
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
	}

	// Create the query
	values := make([]string, len(m.columns))
	for i, _ := range values {
		values[i] = fmt.Sprintf(`$%d`, i+1)
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		m.table,
		strings.Join(m.columns, ", "),
		strings.Join(values, ", "),
	)
	_, err = m.db.Exec(query, &session.Key, &session.Data, &session.Expires)
	// Return nil on error - don't return a session if it wasn't created
	if err != nil {
		return nil, err
	}
	return session, nil
}

// On init:
// * Create a list of valid columns
func init() {
	// Get all the tags
	// TODO Allow for private or unexported fields
	session := &Session{}
	elem := reflect.TypeOf(session).Elem()

	columns := make([]string, elem.NumField())
	for i := 0; i < elem.NumField(); i++ {
		columns[i] = elem.Field(i).Tag.Get("db")
	}

	Sessions = &SessionManager{
		table:   "django_session",
		columns: columns,
	}
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
	var sessionData SessionData
	if err = json.Unmarshal(parts[1], &sessionData); err != nil {
		return nil, BadSessionData
	}

	return &sessionData, nil
}
