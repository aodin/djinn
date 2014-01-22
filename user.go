package djinn

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

var (
	UserDoesNotExist = errors.New("djinn: user does not exist")
	MultipleUsers    = errors.New("djinn: multiple users returned")
	UnusablePassword = errors.New("djinn: user password is unusable")
	UserWithoutId    = errors.New("djinn: user must have an id")
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
	// TODO Add the user manager?
}

func (u *User) String() string {
	return u.Username
}

// Delete the user from the database
func (u *User) Delete() error {
	if u.Id == 0 {
		return UserWithoutId
	}
	// TODO Include a manager object in each user instance
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, Users.table)
	_, err := Users.db.Exec(query, u.Id)
	return err
}

// Update the user
func (u *User) Save() error {
	if u.Id == 0 {
		return UserWithoutId
	}
	// TODO This is a little shady
	// TODO Only update the properties that changed?
	columns := Users.columns[1:]
	values := make([]string, len(columns))
	for i, column := range columns {
		values[i] = fmt.Sprintf(`%s = $%d`, column, i+1)
	}
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, Users.table, strings.Join(values, ", "), len(columns)+1)

	// Build the list of parameters
	elem := reflect.ValueOf(u).Elem()
	parameters := make([]interface{}, len(columns)+1)
	for i := 1; i < elem.NumField(); i++ {
		parameters[i-1] = elem.Field(i).Addr().Interface()
	}
	parameters[len(columns)] = u.Id

	// TODO There should really be a manager object tied to the user struct
	log.Println("djinn update:", query)
	_, err := Users.db.Exec(query, parameters...)
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

// The user manager instance that will be populated on init()
var Users *UserManager

type UserManager struct {
	db      *sql.DB
	table   string
	columns []string
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

// On init:
// * Create a list of valid columns
func init() {
	// Get all the tags
	// TODO Allow for private or unexported fields
	user := &User{}
	elem := reflect.TypeOf(user).Elem()

	columns := make([]string, elem.NumField())
	for i := 0; i < elem.NumField(); i++ {
		columns[i] = elem.Field(i).Tag.Get("db")
	}

	Users = &UserManager{
		table:   "auth_user",
		columns: columns,
	}
}

func (m *UserManager) All() (users []*User, err error) {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(m.columns, ", "), m.table)

	// TODO performance of the interface building versus direct struct scan?
	rows, err := m.db.Query(query)
	if err != nil {
		return
	}
	for rows.Next() {
		user := &User{}
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
	// TODO This is dialect dependent
	// TODO We want the columns except for the id, we know it's first for now
	columns := m.columns[1:]
	values := make([]string, len(columns))
	for i, _ := range values {
		values[i] = fmt.Sprintf(`$%d`, i+1)
	}

	// Build the destination interfaces
	elem := reflect.ValueOf(user).Elem()
	parameters := make([]interface{}, len(columns))
	for i := 1; i < elem.NumField(); i++ {
		parameters[i-1] = elem.Field(i).Addr().Interface()
	}
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING id`, m.table, strings.Join(columns, ", "), strings.Join(values, ", "))

	// Return the new user's id
	log.Println("djinn query:", query)
	err = m.db.QueryRow(query, parameters...).Scan(&user.Id)
	return user, err
}

func (m *UserManager) CreateUser(username, email, password string) (*User, error) {
	return m.createUser(username, email, password, false, false)
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
	user := &User{}

	// Build the WHERE statement
	wheres := make([]string, 0)
	parameters := make([]interface{}, 0)
	paramCount := 0
	for key, value := range values {
		if !m.isValid(key) {
			return nil, errors.New(fmt.Sprintf(`djinn: invalid column %q in user query`, key))
		}
		paramCount += 1
		wheres = append(wheres, fmt.Sprintf(`%s = $%d`, key, paramCount))
		parameters = append(parameters, value)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 2", strings.Join(m.columns, ", "), m.table, strings.Join(wheres, " AND "))

	// Build the destination interfaces
	elem := reflect.ValueOf(user).Elem()
	dest := make([]interface{}, elem.NumField())
	for i := 0; i < elem.NumField(); i++ {
		dest[i] = elem.Field(i).Addr().Interface()
	}

	log.Println("djinn query:", query)
	rows, err := m.db.Query(query, parameters...)
	if err != nil {
		return nil, err
	}

	// One, and only one, result should be returned
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
