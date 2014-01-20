package djinn

import (
	"time"
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
