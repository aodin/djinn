package djinn

import (
	"errors"
	"net/http"
)

var IncorrectPassword = errors.New("djinn: the password was incorrect")

// Read the http.Request object and authenticate a user.
// Returns the User if valid or nil otherwise.
// TODO What to do about possible database and decoding errors?
func Authenticate(req *http.Request) (*User, error) {
	// Get the session cookie
	sessionCookie, err := req.Cookie(config.SessionCookieName)
	if err != nil {
		return nil, err
	}

	// Get the session associated with this key
	session, err := Sessions.Get(sessionCookie.Value)
	if err != nil {
		return nil, err
	}

	// Decode the session data using the salt and secret from config
	sessionData, err := DecodeSessionData(
		[]byte(config.SessionSalt),
		[]byte(config.Secret),
		session.Data,
	)
	if err != nil {
		return nil, err
	}

	// Get the User with the associated Id
	return Users.GetId(sessionData.AuthUserId)
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
	// There must be one, and only one, user returned
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

// Read the session cookie from the request and delete associated session
// from the database if it exists
func Logout(req *http.Request) error {
	// Delete the existing session
	sessionCookie, err := req.Cookie(config.SessionCookieName)
	if err != nil {
		return err
	}

	// Get the session associated with this key
	session, err := Sessions.Get(sessionCookie.Value)
	if err != nil {
		return err
	}

	// TODO Create an anonymous session?
	return session.Delete()
}

// Confirm that the user is logged in or redirect them to the login URL
// TODO Attach the user to the request?
func LoginRequired(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return UserPassesTest(h, func(u *User) bool { return u != nil })
}

// Confirm that the user passes the given test or redirect to the login URL
func UserPassesTest(h func(http.ResponseWriter, *http.Request), test func(*User) bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		user, err := Authenticate(req)
		// TODO 500 status for some errors?
		if err != nil || !test(user) {
			// TODO Set the next header
			http.Redirect(w, req, config.LoginURL, 302)
			return
		}
		h(w, req)
	}
}
