package djinn

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type Values map[string]interface{}

// Wrap the sql.DB to provide access to dialect-specific operations and logging
type DB struct {
	*sql.DB
	dialect Dialect
}

func (d *DB) Query(q string, args ...interface{}) (rows *sql.Rows, err error) {
	before := time.Now()
	rows, err = d.DB.Query(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return
}

func (d *DB) QueryRow(q string, args ...interface{}) *sql.Row {
	before := time.Now()
	row := d.DB.QueryRow(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return row
}

func (d *DB) Exec(q string, args ...interface{}) (sql.Result, error) {
	before := time.Now()
	result, err := d.DB.Exec(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return result, err
}

// Output the elapsed time, query, and arguments
func (d *DB) Log(q string, elapsed time.Duration, args ...interface{}) {
	// TODO Only output if a logger exists
	log.Printf(`(%.3f) %s args=%v`, elapsed.Seconds(), q, args)
}

// Escape table columns and join
func (d *DB) JoinColumns(columns []string) string {
	escaped := make([]string, len(columns))
	for i, column := range columns {
		escaped[i] = fmt.Sprintf(`"%s"`, column)
	}
	return strings.Join(escaped, ", ")
}

func (d *DB) JoinColumnParametersWith(columns []string, sep string, start int) string {
	escaped := make([]string, len(columns))
	for i, column := range columns {
		escaped[i] = fmt.Sprintf(`"%s" = %s`, column, d.dialect.Parameter(i+i))
	}
	return strings.Join(escaped, sep)
}

func (d *DB) JoinColumnParameters(columns []string) string {
	return d.JoinColumnParametersWith(columns, ", ", 0)
}

func (d *DB) BuildParameters(columns []string) string {
	parameters := make([]string, len(columns))
	for i, _ := range columns {
		parameters[i] = d.dialect.Parameter(i)
	}
	return strings.Join(parameters, ", ")
}

// A Dialect must implement all the operations that vary database to database
type Dialect interface {
	Parameter(i int) string
	InsertReturningId(m *Manager, columns []string, params ...interface{}) (int64, error)
}

// TODO A base dialect?

type PostGres struct{}

func (d *PostGres) Parameter(i int) string {
	return fmt.Sprintf(`$%d`, i+1)
}

func (d *PostGres) InsertReturningId(m *Manager, columns []string, params ...interface{}) (id int64, err error) {
	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s) RETURNING %s`,
		m.table,
		m.db.JoinColumns(columns),
		m.db.BuildParameters(columns),
		m.primary,
	)
	err = m.db.QueryRow(query, params...).Scan(&id)
	return id, err
}

type Sqlite3 struct{}

func (d *Sqlite3) Parameter(i int) string {
	return `?`
}

func (d *Sqlite3) InsertReturningId(m *Manager, columns []string, params ...interface{}) (int64, error) {
	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES (%s)`,
		m.table,
		m.db.JoinColumns(columns),
		m.db.BuildParameters(columns),
	)
	result, err := m.db.Exec(query, params...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// The single dialect instance that all managers will embed
// TODO What about multiple databases?
var connection DB

// TODO Bootstrap sequence
var dialects = make(map[string]Dialect)

// TODO register a complete dialect
func RegisterDialect(name string, d Dialect) {
	if d == nil {
		panic("djinn: attempting to register a nil dialect")
	}
	if _, duplicate := dialects[name]; duplicate {
		panic("djinn: RegisterDialect called twice for Dialect " + name)
	}
	dialects[name] = d
}

func GetDialect(name string) (Dialect, error) {
	d, ok := dialects[name]
	if !ok {
		return nil, fmt.Errorf("djinn: unknown dialect %s (did you remember to import it?)", name)
	}
	return d, nil
}

func init() {
	RegisterDialect("sqlite3", &Sqlite3{})
	RegisterDialect("postgres", &PostGres{})
}

func Connect(driver, credentials string) (*DB, error) {
	db, err := sql.Open(driver, credentials)
	if err != nil {
		return nil, err
	}
	// Determine which parameter building should be used
	dialect, err := GetDialect(driver)
	if err != nil {
		return nil, err
	}
	connection = DB{DB: db, dialect: dialect}
	return &connection, nil
}
