package djinn

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type Values map[string]interface{}

type Dialect struct {
	*sql.DB
	parameters ParameterBuilder
}

func (d *Dialect) Query(q string, args ...interface{}) (rows *sql.Rows, err error) {
	before := time.Now()
	rows, err = d.DB.Query(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return
}

func (d *Dialect) QueryRow(q string, args ...interface{}) *sql.Row {
	before := time.Now()
	row := d.DB.QueryRow(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return row
}

func (d *Dialect) Exec(q string, args ...interface{}) (sql.Result, error) {
	before := time.Now()
	result, err := d.DB.Exec(q, args...)
	d.Log(q, time.Now().Sub(before), args)
	return result, err
}

// Output the elapsed time, query, and arguments
func (d *Dialect) Log(q string, elapsed time.Duration, args ...interface{}) {
	// TODO Only output if a logger exists
	log.Printf(`(%.3f) %s args=%v`, elapsed.Seconds(), q, args)
}

// Escape table columns and join
func (d *Dialect) JoinColumns(columns []string) string {
	escaped := make([]string, len(columns))
	for i, column := range columns {
		escaped[i] = fmt.Sprintf(`"%s"`, column)
	}
	return strings.Join(escaped, ", ")
}

func (d *Dialect) JoinColumnParametersWith(columns []string, sep string, start int) string {
	escaped := make([]string, len(columns))
	for i, column := range columns {
		escaped[i] = fmt.Sprintf(`"%s" = %s`, column, d.parameters.Build(i+i))
	}
	return strings.Join(escaped, sep)
}

func (d *Dialect) JoinColumnParameters(columns []string) string {
	return d.JoinColumnParametersWith(columns, ", ", 0)
}

func (d *Dialect) BuildParameters(columns []string) string {
	parameters := make([]string, len(columns))
	for i, _ := range columns {
		parameters[i] = d.parameters.Build(i)
	}
	return strings.Join(parameters, ", ")
}

type ParameterBuilder interface {
	Build(int) string
}

type PostGresBuilder struct{}

func (b *PostGresBuilder) Build(i int) string {
	return fmt.Sprintf(`$%d`, i+1)
}

type Sqlite3Builder struct{}

func (b *Sqlite3Builder) Build(i int) string {
	return `?`
}

// The single dialect instance that all managers will embed
// TODO What about multiple databases?
var dialect Dialect

// TODO Bootstrap sequence
var dialects = make(map[string]ParameterBuilder)

// TODO register a complete dialect
func RegisterDialect(name string, d ParameterBuilder) {
	if d == nil {
		panic("djinn: attempting to register a nil dialect")
	}
	if _, duplicate := dialects[name]; duplicate {
		panic("djinn: RegisterDialect called twice for Dialect " + name)
	}
	dialects[name] = d
}

func GetDialect(name string) (ParameterBuilder, error) {
	d, ok := dialects[name]
	if !ok {
		return nil, fmt.Errorf("djinn: unknown dialect %s (did you remember to import it?)", name)
	}
	return d, nil
}

func init() {
	RegisterDialect("sqlite3", &Sqlite3Builder{})
	RegisterDialect("postgres", &PostGresBuilder{})
}

func Connect(driver, credentials string) (*Dialect, error) {
	db, err := sql.Open(driver, credentials)
	if err != nil {
		return nil, err
	}
	// Determine which parameter building should be used
	parameters, err := GetDialect(driver)
	if err != nil {
		return nil, err
	}
	dialect = Dialect{DB: db, parameters: parameters}
	return &dialect, nil
}
