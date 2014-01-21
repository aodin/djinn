package djinn

import (
	_ "github.com/lib/pq"
	"testing"
)

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

func TestUser(t *testing.T) {
	// Connect to the database
	db, err := Connect("postgres", "host=localhost port=5432 dbname=djangoex user=postgres password=gotest")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	user, err := Users.GetId(1)
	if err != nil {
		t.Fatal(err)
	}

	ExpectString(t, user.Username, "admin")
}
