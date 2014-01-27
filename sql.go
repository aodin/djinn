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

var CREATE_DJANGO_AUTH = `BEGIN;
CREATE TABLE "auth_permission" (
    "id" serial NOT NULL PRIMARY KEY,
    "name" varchar(50) NOT NULL,
    "content_type_id" integer NOT NULL REFERENCES "django_content_type" ("id") DEFERRABLE INITIALLY DEFERRED,
    "codename" varchar(100) NOT NULL,
    UNIQUE ("content_type_id", "codename")
)
;
CREATE TABLE "auth_group_permissions" (
    "id" serial NOT NULL PRIMARY KEY,
    "group_id" integer NOT NULL,
    "permission_id" integer NOT NULL REFERENCES "auth_permission" ("id") DEFERRABLE INITIALLY DEFERRED,
    UNIQUE ("group_id", "permission_id")
)
;
CREATE TABLE "auth_group" (
    "id" serial NOT NULL PRIMARY KEY,
    "name" varchar(80) NOT NULL UNIQUE
)
;
ALTER TABLE "auth_group_permissions" ADD CONSTRAINT "group_id_refs_id_f4b32aac" FOREIGN KEY ("group_id") REFERENCES "auth_group" ("id") DEFERRABLE INITIALLY DEFERRED;
CREATE TABLE "auth_user_groups" (
    "id" serial NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL,
    "group_id" integer NOT NULL REFERENCES "auth_group" ("id") DEFERRABLE INITIALLY DEFERRED,
    UNIQUE ("user_id", "group_id")
)
;
CREATE TABLE "auth_user_user_permissions" (
    "id" serial NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL,
    "permission_id" integer NOT NULL REFERENCES "auth_permission" ("id") DEFERRABLE INITIALLY DEFERRED,
    UNIQUE ("user_id", "permission_id")
)
;
CREATE TABLE "auth_user" (
    "id" serial NOT NULL PRIMARY KEY,
    "password" varchar(128) NOT NULL,
    "last_login" timestamp with time zone NOT NULL,
    "is_superuser" boolean NOT NULL,
    "username" varchar(30) NOT NULL UNIQUE,
    "first_name" varchar(30) NOT NULL,
    "last_name" varchar(30) NOT NULL,
    "email" varchar(75) NOT NULL,
    "is_staff" boolean NOT NULL,
    "is_active" boolean NOT NULL,
    "date_joined" timestamp with time zone NOT NULL
)
;
ALTER TABLE "auth_user_groups" ADD CONSTRAINT "user_id_refs_id_40c41112" FOREIGN KEY ("user_id") REFERENCES "auth_user" ("id") DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE "auth_user_user_permissions" ADD CONSTRAINT "user_id_refs_id_4dc23c39" FOREIGN KEY ("user_id") REFERENCES "auth_user" ("id") DEFERRABLE INITIALLY DEFERRED;
CREATE INDEX "auth_permission_content_type_id" ON "auth_permission" ("content_type_id");
CREATE INDEX "auth_group_permissions_group_id" ON "auth_group_permissions" ("group_id");
CREATE INDEX "auth_group_permissions_permission_id" ON "auth_group_permissions" ("permission_id");
CREATE INDEX "auth_group_name_like" ON "auth_group" ("name" varchar_pattern_ops);
CREATE INDEX "auth_user_groups_user_id" ON "auth_user_groups" ("user_id");
CREATE INDEX "auth_user_groups_group_id" ON "auth_user_groups" ("group_id");
CREATE INDEX "auth_user_user_permissions_user_id" ON "auth_user_user_permissions" ("user_id");
CREATE INDEX "auth_user_user_permissions_permission_id" ON "auth_user_user_permissions" ("permission_id");
CREATE INDEX "auth_user_username_like" ON "auth_user" ("username" varchar_pattern_ops);

COMMIT;`

var CREATE_DJANGO_SESSION = `BEGIN;
CREATE TABLE "django_session" (
    "session_key" varchar(40) NOT NULL PRIMARY KEY,
    "session_data" text NOT NULL,
    "expire_date" timestamp with time zone NOT NULL
)
;
CREATE INDEX "django_session_session_key_like" ON "django_session" ("session_key" varchar_pattern_ops);
CREATE INDEX "django_session_expire_date" ON "django_session" ("expire_date");

COMMIT;`
