package djinn

import (
	"errors"
	"net/http"
)

var IncorrectPassword = errors.New("djinn: the password was incorrect")

// Read the http.Request object and authenticate a user.
// Returns the User if valid or nil otherwise.
// TODO What to do about possible database and decoding errors?
func Authenticate(req *http.Request) *User {
	// Get the session cookie
	sessionCookie, err := req.Cookie(config.SessionCookieName)
	if err != nil {
		return nil
	}

	// Get the session associated with this key
	session, err := Sessions.Get(sessionCookie.Value)
	if err != nil {
		return nil
	}

	// Decode the session data using the salt and secret from config
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

// Cookies must be written before any data.
func SetSessionCookie(w http.ResponseWriter, session *Session) {
	// Create the cookie
	cookie := &http.Cookie{
		Name:     config.SessionCookieName,
		Value:    session.Key,
		Path:     config.SessionCookiePath,
		Domain:   config.SessionCookieDomain,
		Expires:  session.Expires,
		HttpOnly: config.SessionCookieHttpOnly,
		Secure:   config.SessionCookieSecure,
	}
	http.SetCookie(w, cookie)
}

// Read the http.Request object and Log in a user.
// Returns the User and writes a session cookie to http.ResponseWriter.
// This function must be called before anything is written to the response.
func Login(w http.ResponseWriter, req *http.Request) (*User, error) {
	// TODO Custom username and password fields
	username := req.FormValue("username")
	password := req.FormValue("password")

	// Get the user with this username
	// There must only be one user returned
	user, err := Users.Get(Values{"username": username})
	if err != nil {
		return nil, err
	}

	// Do a constant time comparison of passwords
	valid, err := user.CheckPassword(password)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, IncorrectPassword
	}

	// Create a new session
	session, err := Sessions.Create(user.Id)
	if err != nil {
		return nil, err
	}

	// Set the cookie and return the user
	SetSessionCookie(w, session)
	return user, nil
}
