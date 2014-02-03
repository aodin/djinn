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

func expectDuration(t *testing.T, a time.Duration, b string) {
	d, err := time.ParseDuration(b)
	if err != nil {
		t.Error(err)
	}
	if a != d {
		t.Errorf("Unexpected duration: %v != %v", a, d)
	}
}

var sqliteUserSchema = `CREATE TABLE "auth_user" (
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
);`

func TestUsers(t *testing.T) {
	// Set the default hasher to MD5 for fast testing
	// TODO Reset after testing is complete
	config.PasswordHasher = "md5"

	// Start an in-memory sql database for testing
	db := createSqliteTestSchema(t, sqliteUserSchema)
	defer db.Close()

	// Create a user
	user, err := Users.CreateUser("client", "", "client")
	if err != nil {
		t.Fatal(err)
	}
	expectString(t, user.Username, "client")
	// TODO Do we have an expected id?
	expectInt64(t, user.Id, 1)

	// Get a user that exists by Id
	client, err := Users.GetId(user.Id)
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

	// Create another User
	_, err = Users.CreateSuperuser("admin", "admin@example.com", "admin")
	if err != nil {
		t.Fatal(err)
	}

	users, err := Users.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("Unexpected length of Users.All(): %s != 2", len(users))
	}
}
