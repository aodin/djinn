package djinn

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

var sqliteGroupSchema = `CREATE TABLE "auth_group" (
    "id" integer NOT NULL PRIMARY KEY,
    "name" varchar(80) NOT NULL UNIQUE
);`

func TestGroups(t *testing.T) {
	// Start an in-memory sql database for testing
	db := createSqliteTestSchema(t, sqliteGroupSchema)
	defer db.Close()

	// Create a group
	group, err := Groups.Create("example")
	if err != nil {
		t.Fatal(err)
	}

	// TODO Do we expect a certain id?
	expectInt64(t, group.Id, 1)
	expectString(t, group.Name, "example")

	// Select it by id
	group, err = Groups.GetId(1)
	if err != nil {
		t.Fatal(err)
	}
	expectString(t, group.Name, "example")

	// And name
	group, err = Groups.Get(Values{"name": "example"})
	expectInt64(t, group.Id, 1)

	// Add a few more groups
	_, err = Groups.Create("example2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Groups.Create("TEST")
	if err != nil {
		t.Fatal(err)
	}

	// Select all
	groups, err := Groups.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 3 {
		t.Errorf("Unexpected length of groups: %d != 3", len(groups))
	}

	// Delete a group
	err = groups[2].Delete()
	if err != nil {
		t.Fatal(err)
	}

	// Update another

	// Select with a filter

}
