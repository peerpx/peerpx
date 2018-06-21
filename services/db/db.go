package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/golang-migrate/migrate/source/github"
	"github.com/labstack/gommon/log"

	"os"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var (
	// ErrNotInitialized (db == nil)
	ErrNotInitialized = errors.New("db: service not initialized")
)

var db *sqlx.DB
var Mock sqlmock.Sqlmock

// InitDatabase init database handler
func InitDatabase(driverName, dataSourceName string) (err error) {
	// migrate DB
	if err := migrateDb(); err != nil {
		return err
	}
	db, err = sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func migrateDb() error {
	var m *migrate.Migrate
	var err error
	// check for local file
	if _, err = os.Stat("./sql2"); os.IsNotExist(err) {
		// migrate from github

		m, err = migrate.New("github://toorop:a8ded4740bc467f6203f85f6ffe9c0cdf25515c7@peerpx/peerpx/cmd/server/dist/sql", "sqlite3://peerpx.db")
		log.Infof("migrate from github %v %v", m, err)
	} else {
		// from local file
		m, err = migrate.New("file://sql", "sqlite3://peerpx.db")
	}
	if err != nil {
		return err
	}
	defer m.Close()
	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
	}
	return err
}

// InitMockedDatabase initialize a mocked DB for testing purpose
// https://github.com/jmoiron/sqlx/issues/204
func InitMockedDatabase() {
	if db != nil && Mock != nil {
		return
	}
	var err error
	var mockDB *sql.DB
	mockDB, Mock, err = sqlmock.New()
	if err != nil {
		panic(fmt.Sprintf("slqmock initialization failed: %v", err))
	}
	db = sqlx.NewDb(mockDB, "sqlmock")
}

func BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.BeginTxx(ctx, opts)
}

func Beginx() (*sqlx.Tx, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.Beginx()
}

func BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	if db == nil {
		return "", nil, ErrNotInitialized
	}
	return db.BindNamed(query, arg)
}

func DriverName() string {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.DriverName()
}

func Get(dest interface{}, query string, args ...interface{}) error {
	if db == nil {
		return ErrNotInitialized
	}
	return db.Get(dest, query, args...)
}

func GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if db == nil {
		return ErrNotInitialized
	}
	return db.GetContext(ctx, dest, query, args...)
}

func MapperFunc(mf func(string) string) {
	if db == nil {
		panic(ErrNotInitialized)
	}
	db.MapperFunc(mf)
}

func MustBegin() *sqlx.Tx {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.MustBegin()
}

func MustBeginTx(ctx context.Context, opts *sql.TxOptions) *sqlx.Tx {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.MustBeginTx(ctx, opts)
}

func MustExec(query string, args ...interface{}) sql.Result {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.MustExec(query, args...)
}

func MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.MustExecContext(ctx, query, args...)
}

func NamedExec(query string, arg interface{}) (sql.Result, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.NamedExec(query, arg)
}

func NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.NamedExecContext(ctx, query, arg)
}

func NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.NamedQuery(query, arg)
}

func NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.NamedQueryContext(ctx, query, arg)
}

func PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.PrepareNamed(query)
}

func PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.PrepareNamedContext(ctx, query)
}

func Preparex(query string) (*sqlx.Stmt, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.Preparex(query)
}

func PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.PreparexContext(ctx, query)
}

func QueryRowx(query string, args ...interface{}) *sqlx.Row {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.QueryRowx(query, args...)
}

func QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.QueryRowxContext(ctx, query, args...)
}

func Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.Queryx(query, args...)
}

func QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db.QueryxContext(ctx, query, args...)
}

func Rebind(query string) string {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.Rebind(query)
}
func Select(dest interface{}, query string, args ...interface{}) error {
	if db == nil {
		return ErrNotInitialized
	}
	return db.Select(dest, query, args...)
}

func SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if db == nil {
		return ErrNotInitialized
	}
	return db.SelectContext(ctx, dest, query, args...)
}

func Unsafe() *sqlx.DB {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db.Unsafe()
}
