package djinn

import (
	"testing"
)

func createSqliteTestSchema(t *testing.T, schemas ...string) *DB {
	db, err := Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create the given schemas
	// Use the internal db connection so the schema queries are not logged
	// TODO Error if no schemas were given?
	for _, schema := range schemas {
		_, err = db.DB.Exec(schema)
		if err != nil {
			t.Fatal(err)
		}
	}
	return db
}
