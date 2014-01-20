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
}

func (u *User) String() string {
	return u.Username
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

// The specific manager instance that will be populated on init()
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
	// TODO Where should the default hasher be set?
	user := &User{
		Username:    username,
		Password:    MakePassword(defaultHasher, password),
		Email:       email,
		IsStaff:     is_staff,
		IsSuperuser: is_superuser,
	}

	// Build a list of parameters
	// TODO This is dialect dependent
	values := make([]string, len(m.columns))
	for i, _ := range values {
		values[i] = fmt.Sprintf(`$%d`, i+1)
	}

	// Build the destination interfaces
	elem := reflect.ValueOf(user).Elem()
	parameters := make([]interface{}, elem.NumField())
	for i := 0; i < elem.NumField(); i++ {
		parameters[i] = elem.Field(i).Addr().Interface()
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING id`, strings.Join(m.columns, ", "), m.table, strings.Join(values, ", "))

	// Return the new user's id
	err := m.db.QueryRow(query, parameters...).Scan(&user.Id)
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
