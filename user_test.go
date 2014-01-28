package djinn

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
)

// TODO A better place for testing functions?
func expectString(t *testing.T, a, b string) {
	if a != b {
		t.Errorf("Unexpected string: %s != %s", a, b)
	}
}

func expectInt64(t *testing.T, a, b int64) {
	if a != b {
		t.Errorf("Unexpected integer: %d != %d", a, b)
	}
}

func expectInt(t *testing.T, a, b int) {
	if a != b {
		t.Errorf("Unexpected integer: %d != %d", a, b)
	}
}

var sqliteUserSchema = `
CREATE TABLE "auth_user" (
    "id" integer NOT NULL PRIMARY KEY,
    "password" varchar(128) NOT NULL,
    "last_login" datetime NOT NULL,
    "is_superuser" bool NOT NULL,
    "username" varchar(30) NOT NULL UNIQUE,
    "first_name" varchar(30) NOT NULL,
    "last_name" varchar(30) NOT NULL,
    "email" varchar(75) NOT NULL,
    "is_staff" bool NOT NULL,
    "is_active" bool NOT NULL,
    "date_joined" datetime NOT NULL
)
;
`

var sqliteInsertUser = `
 INSERT INTO "auth_user" ("username", "password", "first_name", "last_name", "email", "is_active", "is_staff", "is_superuser", "date_joined", "last_login") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

var exc = &User{
	Id:          1,
	Username:    "client",
	Password:    "pbkdf2_sha256$12000$vfl5YUMEhry5$v4CCOHbNUyzku3s27rh1B3UIoqNzYoG0jV9CHpUHXAQ=", // "client"
	FirstName:   "",
	LastName:    "",
	Email:       "",
	IsActive:    true,
	IsStaff:     false,
	IsSuperuser: false,
	DateJoined:  time.Now(),
	LastLogin:   time.Now(),
	manager:     Users,
}

func TestUsers(t *testing.T) {
	// Start an in-memory sql database for testing
	db := createSqliteTestSchema(t)
	defer db.Close()

	// Create a user
	// TODO Need to fix "RETURNING" dialect specific syntax
	// user, err := Users.CreateUser("client", "", "client")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// expectString(t, user.Username, "client")

	// Get a user that exists by Id
	client, err := Users.GetId(1)
	if err != nil {
		t.Fatal(err)
	}
	expectString(t, client.Username, "client")

	// Get a user that exists by Name
	client, err = Users.Get(Values{"username": "client"})
	if err != nil {
		t.Fatal(err)
	}
	expectInt64(t, client.Id, 1)

	// Query mutliple attributes
	client, err = Users.Get(Values{"id": 1, "username": "client"})
	if err != nil {
		t.Fatal(err)
	}
	expectInt64(t, client.Id, 1)

	// Get a user that does not exist
	client, err = Users.GetId(2)
	if err != UserDoesNotExist {
		t.Error("Expected a UserDoesNotExist error, but one did not occur")
	}

	// Attempt a query by an attribute that does not exist
	_, err = Users.Get(Values{"sparkles": 23})
	if err == nil {
		t.Error("Expected an error from an invalid attribute in Users.Get(), but one did not occur")
	}
}
