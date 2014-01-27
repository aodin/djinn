package djinn

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
)

// TODO A better place for testing functions?
func ExpectString(t *testing.T, a, b string) {
	if a != b {
		t.Errorf("Unexpected string: %s != %s", a, b)
	}
}

func ExpectInt(t *testing.T, a, b int64) {
	if a != b {
		t.Errorf("Unexpected integer: %d != %d", a, b)
	}
}

var bootstrapSqlite3 = `
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

var sqlite3user = `
 INSERT INTO "auth_user" ("username", "password", "first_name", "last_name", "email", "is_active", "is_staff", "is_superuser", "date_joined", "last_login") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

var exc = &User{
	Id:          1,
	Username:    "client",
	Password:    "pbkdf2_sha256$12000$vfl5YUMEhry5$v4CCOHbNUyzku3s27rh1B3UIoqNzYoG0jV9CHpUHXAQ=",
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

func TestUser(t *testing.T) {
	// Start an in-memory sql database for testing
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Add the Users table
	_, err = db.Exec(bootstrapSqlite3)
	if err != nil {
		t.Fatal(err)
	}

	// TODO Manually insert a user for now
	_, err = db.Exec(sqlite3user, exc.Username, exc.Password, exc.FirstName, exc.LastName, exc.Email, exc.IsActive, exc.IsStaff, exc.IsSuperuser, exc.DateJoined, exc.LastLogin)
	if err != nil {
		t.Fatal(err)
	}

	// Create a user
	// TODO Need to fix "RETURNING" dialect specific syntax
	// user, err := Users.CreateUser("client", "", "client")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// ExpectString(t, user.Username, "client")

	client, err := Users.GetId(1)
	if err != nil {
		t.Fatal(err)
	}
	ExpectString(t, client.Username, "client")
}
