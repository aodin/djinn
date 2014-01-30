package djinn

import (
	"errors"
	"fmt"
)

var (
	GroupDoesNotExist = errors.New("djinn: group does not exist")
	MultipleGroups    = errors.New("djinn: group does not exist")
)

// auth_group
type Group struct {
	Id      int64  `db:"id"`
	Name    string `db:"name"`
	manager *GroupManager
}

// TODO Many to many link to permissions

func (group *Group) String() string {
	return group.Name
}

func (g *Group) Delete() error {
	return g.manager.deleteOne(g)
}

type GroupManager struct {
	*Manager
}

var Groups = &GroupManager{
	&Manager{
		db:      &connection,
		table:   "auth_group",
		columns: []string{"id", "name"},
		primary: "id",
	},
}

func (m *GroupManager) All() (groups []*Group, err error) {
	query := fmt.Sprintf(
		`SELECT %s FROM "%s"`,
		m.db.JoinColumns(m.columns),
		m.table,
	)
	rows, err := m.db.Query(query)
	if err != nil {
		return
	}
	for rows.Next() {
		group := &Group{
			manager: m,
		}
		if err = rows.Scan(&group.Id, &group.Name); err != nil {
			return
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func (m *GroupManager) deleteOne(group *Group) error {
	query := fmt.Sprintf(
		`DELETE FROM "%s" WHERE "%s" = %s`,
		m.table,
		m.primary,
		m.db.dialect.Parameter(0),
	)
	// TODO Save the result of the last operation to the manager?
	_, err := m.db.Exec(query, group.Id)
	return err
}

func (m *GroupManager) Create(name string) (*Group, error) {
	group := &Group{
		Name:    name,
		manager: m,
	}
	// INSERT with an auto-increment id varies between dialects
	id, err := m.db.dialect.InsertReturningId(m.Manager, []string{"name"}, &group.Name)
	if err != nil {
		return nil, err
	}
	group.Id = id
	return group, err
}

// Users.Get() behavior (disregarding database or attribute errors):
// * No results:       (nil, GroupDoesNotExist)
// * One result:       (<user>, nil)
// * Multiple results: (nil, MultipleGroups)
func (m *GroupManager) GetId(id int64) (*Group, error) {
	return m.Get(Values{"id": id})
}

func (m *GroupManager) Get(values Values) (*Group, error) {
	// TODO There must be a database connection and at least one value

	// Build the WHERE statement
	// These must equal the values given or the function returns an error
	// TODO Generalize the building of WHERE statements
	parameters := make([]interface{}, len(values))
	valid := make([]string, len(values))

	index := 0
	for key, value := range values {
		if !m.isValid(key) {
			return nil, fmt.Errorf(`djinn: invalid column %q in Groups.Get()`, key)
		}
		valid[index] = key
		parameters[index] = value
		index += 1
	}
	query := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE %s LIMIT 2`,
		m.db.JoinColumns(m.columns),
		m.table,
		m.db.JoinColumnParametersWith(valid, " AND ", 0),
	)

	// Destination for queried Group
	group := &Group{
		manager: m,
	}

	rows, err := m.db.Query(query, parameters...)
	if err != nil {
		return nil, err
	}

	// One, and only one result should be returned
	if !rows.Next() {
		return nil, GroupDoesNotExist
	}
	if err := rows.Scan(&group.Id, &group.Name); err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, MultipleGroups
	}
	return group, nil
}
