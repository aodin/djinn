package djinn

import (
	"net/http"
)

// Read the http.Request object and authenticate a user.
// Returns the User if valid or nil otherwise.
// TODO What to do about possible database and decoding errors?
func Authenticate(req *http.Request) *User {
	// Get the session cookie
	sessionCookie, err := req.Cookie(config.SessionCookieName)
	if err != nil {
		return nil
	}
	sessionKey := sessionCookie.Value

	// Get the session associated with this id
	session, err := Sessions.Get(sessionKey)
	if err != nil {
		return nil
	}

	sessionData, err := DecodeSessionData(
		[]byte(config.SessionSalt),
		[]byte(config.Secret),
		session.Data,
	)
	if err != nil {
		return nil
	}

	// Get the User with the associated Id
	user, err := Users.GetId(sessionData.AuthUserId)
	if err != nil {
		return nil
	}
	return user
}

// Read the http.Request object and Log in a user.
// Returns the User and writes a session cookie to http.ResponseWriter.
// This function must be called before anything is written to the response.
// TODO What to do about possible database and decoding errors?
func Login(w http.ResponseWriter, req *http.Request) *User {
	return nil
}
