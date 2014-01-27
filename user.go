package djinn

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	UserDoesNotExist = errors.New("djinn: user does not exist")
	MultipleUsers    = errors.New("djinn: multiple users returned")
	UnusablePassword = errors.New("djinn: user password is unusable")
)

// auth_user
type User struct {
	Id          int64     `db:"id"`
	Username    string    `db:"username"`
	Password    string    `db:"password"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Email       string    `db:"email"`
	IsActive    bool      `db:"is_active"`
	IsStaff     bool      `db:"is_staff"`
	IsSuperuser bool      `db:"is_superuser"`
	DateJoined  time.Time `db:"date_joined"`
	LastLogin   time.Time `db:"last_login"`
	manager     *UserManager
}

func (u *User) String() string {
	return u.Username
}

// Delete the user from the database
func (u *User) Delete() error {
	// TODO There must be a non-nil manager and database connection
	query := fmt.Sprintf(
		`DELETE FROM "%s" WHERE "%s" = %s`,
		u.manager.table,
		u.manager.primary,
		u.manager.db.parameters.Build(0),
	)
	_, err := u.manager.db.Exec(query, u.Id)
	return err
}

// Update the user
func (u *User) Save() error {
	// TODO There must be a non-nil manager and database connection
	// TODO Only update the properties that changed?
	columns := Users.columns[1:]
	query := fmt.Sprintf(
		`UPDATE "%s" SET %s WHERE "%s" = %s`,
		u.manager.table,
		u.manager.db.JoinColumnParameters(columns),
		u.manager.primary,
		u.manager.db.parameters.Build(len(columns)),
	)

	// Build the list of parameters
	elem := reflect.ValueOf(u).Elem()
	parameters := make([]interface{}, len(columns)+1)
	for i := 1; i < elem.NumField(); i++ {
		parameters[i-1] = elem.Field(i).Addr().Interface()
	}
	parameters[len(columns)] = u.Id

	_, err := u.manager.db.Exec(query, parameters...)
	return err
}

func (u *User) CheckPassword(password string) (bool, error) {
	// TODO There is a redundant split
	// Determine the type of hasher
	parts := strings.SplitN(u.Password, "$", 2)
	if len(parts) != 2 {
		return false, UnusablePassword
	}
	hasher, err := GetHasher(parts[0])
	if err != nil {
		return false, err
	}
	return CheckPassword(hasher, password, u.Password), nil
}

type UserManager struct {
	db      *Dialect
	table   string
	columns []string
	primary string
}

// Build columns and primary keys dynamically - on init?
var Users = &UserManager{
	db:      &dialect,
	table:   "auth_user",
	columns: []string{"id", "username", "password", "first_name", "last_name", "email", "is_active", "is_staff", "is_superuser", "date_joined", "last_login"},
	primary: "id",
}

// TODO A generalized isValid method for all managers
func (m *UserManager) isValid(column string) bool {
	for _, col := range m.columns {
		if column == col {
			return true
		}
	}
	return false
}

func (m *UserManager) All() (users []*User, err error) {
	query := fmt.Sprintf(
		`SELECT %s FROM "%s"`,
		m.db.JoinColumns(m.columns),
		m.table,
	)

	// TODO performance of the interface building versus direct struct scan?
	rows, err := m.db.Query(query)
	if err != nil {
		return
	}
	for rows.Next() {
		user := &User{
			manager: m,
		}
		elem := reflect.ValueOf(user).Elem()
		dest := make([]interface{}, elem.NumField())
		for i := 0; i < elem.NumField(); i++ {
			dest[i] = elem.Field(i).Addr().Interface()
		}

		if err = rows.Scan(dest...); err != nil {
			return
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func (m *UserManager) createUser(username, email, password string, is_staff, is_superuser bool) (*User, error) {
	// Prepare the user
	// TODO Default values are tricky because Go users nil initialization
	now := time.Now()

	// Get the default password hashing algorithm
	defaultHasher, err := GetHasher(config.PasswordHasher)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:    username,
		Password:    MakePassword(defaultHasher, password),
		Email:       email,
		IsSuperuser: is_superuser,
		IsStaff:     is_staff,
		IsActive:    true,
		DateJoined:  now,
		LastLogin:   now,
	}

	// Build a list of parameters
	// TODO We want the columns except for the id, we know it's first for now
	columns := m.columns[1:]

	// Build the destination interfaces

	// Build the destination interfaces
	elem := reflect.ValueOf(user).Elem()
	tags := reflect.TypeOf(user).Elem()
	dest := make([]interface{}, 0)
	// Start at 1 to skip the id
	for i := 1; i < elem.NumField(); i++ {
		if tags.Field(i).Tag.Get("db") != "" {
			dest = append(dest, elem.Field(i).Addr().Interface())
		}
	}

	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s) RETURNING %s`,
		m.table,
		m.db.JoinColumns(columns),
		m.db.BuildParameters(columns),
		m.primary,
	)

	// Return the new user's id
	err = m.db.QueryRow(query, dest...).Scan(&user.Id)
	return user, err
}

func (m *UserManager) CreateUser(username, email, password string) (*User, error) {
	return m.createUser(username, email, password, false, false)
}

func (m *UserManager) CreateStaff(username, email, password string) (*User, error) {
	return m.createUser(username, email, password, true, false)
}

func (m *UserManager) CreateSuperuser(username, email, password string) (*User, error) {
	return m.createUser(username, email, password, true, true)
}

// func (m *UserManager) Filter(values Values) (users []*User, err error) {
// }

func (m *UserManager) GetId(id int64) (*User, error) {
	return m.Get(Values{"id": id})
}

func (m *UserManager) Get(values Values) (*User, error) {
	// TODO There must be a database connection and at least one value
	user := &User{
		manager: m,
	}

	// Build the WHERE statement
	// These must equal the values given or the function returns an error
	parameters := make([]interface{}, len(values))
	valid := make([]string, len(values))

	index := 0
	for key, value := range values {
		if !m.isValid(key) {
			return nil, fmt.Errorf(`djinn: invalid column %q in user query`, key)
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

	// Build the destination interfaces
	elem := reflect.ValueOf(user).Elem()
	tags := reflect.TypeOf(user).Elem()
	dest := make([]interface{}, 0)
	for i := 0; i < elem.NumField(); i++ {
		if tags.Field(i).Tag.Get("db") != "" {
			dest = append(dest, elem.Field(i).Addr().Interface())
		}
	}

	rows, err := m.db.Query(query, parameters...)
	if err != nil {
		return nil, err
	}

	// One, and only one result should be returned
	if !rows.Next() {
		return nil, UserDoesNotExist
	}
	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, MultipleUsers
	}
	return user, nil
}
