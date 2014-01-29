package djinn

import (
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
)

var doNotFollow = errors.New("djinn: do not follow redirects")

// Behavior:
// * GET without a session:              401 Unauthorized
// * GET with a session:                 200 OK
// * POST with correct credentials:      302 Found
// * POST with incorrect credentials:    400
// * All other errors
func loginTestHander(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
	if r.Method == "POST" {
		_, err := Login(w, r)
		if err == IncorrectPassword || err == UserDoesNotExist {
			// Bad credentials
			http.Error(w, "Improper credentials", 400)
			return
		}
		// All other errors are the server's fault
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// By default, the client will follow redirects
		http.Redirect(w, r, "/redirect", 302)
		return
	}
	_, err := Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	w.Write([]byte(`200`))
}

func createSqliteTestSchema(t *testing.T) *DB {
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create the Users schema
	// We use the internal db connection so the schema queries are not logged
	_, err = db.DB.Exec(sqliteUserSchema)
	if err != nil {
		t.Fatal(err)
	}

	// Create the sessions schema
	_, err = db.DB.Exec(sqliteSessionSchema)
	if err != nil {
		t.Fatal(err)
	}

	// TODO Manually insert a user for now
	// TODO This should not be part of the schema set up
	_, err = db.DB.Exec(sqliteInsertUser, exc.Username, exc.Password, exc.FirstName, exc.LastName, exc.Email, exc.IsActive, exc.IsStaff, exc.IsSuperuser, exc.DateJoined, exc.LastLogin)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLogin(t *testing.T) {
	// Set the secret or the session decode will use the default ""
	// TODO Common testing configuration
	secret := `xsy!9deorcwbk!&=u33!ixik-r9c1@sf6tz0jnb*ce9ipe)e&m`
	SetSecret(secret)

	// Start an in-memory sqlite database
	db := createSqliteTestSchema(t)
	defer db.Close()

	// Start the login test server
	ts := httptest.NewServer(http.HandlerFunc(loginTestHander))
	defer ts.Close()

	// Test a login page
	// A GET without a session should return 401 Unauthorized
	response, err := http.Get(ts.URL + "/unauthorized")
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, response.StatusCode, 401)

	// A POST should only return 302 on successful login and 400 for
	// bad credentials
	response, err = http.PostForm(ts.URL+"/bad", url.Values{"username": {"client"}, "password": {"bad"}})
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, response.StatusCode, 400)

	// Use a custom client to control redirect policy and save cookies
	ignoreRedirects := func(r *http.Request, via []*http.Request) error {
		return doNotFollow
	}
	cjar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		CheckRedirect: ignoreRedirects,
		Jar:           cjar,
	}
	response, err = client.PostForm(ts.URL+"/login", url.Values{"username": {"client"}, "password": {"client"}})
	// TODO This should be true: err.Err != doNotFollow
	// But since the error interface is returned, how do you test?
	// Assert the type and check err.Err?
	// if err != nil &&  {
	// 	t.Fatal(err)
	// }
	expectInt(t, response.StatusCode, 302)

	// Get the session cookie from the response and set it for the next request
	// cookies := response.Cookies()

	// TODO cookiejar expects a url.URL but the test server URL is a string
	testURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	// Get the cookies from the custom client cookie jar
	cookies := client.Jar.Cookies(testURL)
	if len(cookies) != 1 {
		t.Fatalf("Unexpected length of login cookies: %d != %d", len(cookies), 1)
	}
	expectString(t, cookies[0].Name, config.SessionCookieName)

	// TODO Test the database entry?

	// There should now be a valid session
	response, err = client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, response.StatusCode, 200)
}
