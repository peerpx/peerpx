package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// DB is the database connector
var DB *gorm.DB

// InitDB initialize database connector
// todo: get options/type from viper
func InitDB() (err error) {
	DB, err = gorm.Open("sqlite3", "peerpx.db")
	return err
}

// InitMockedDB is for tests
func InitMockedDB(dsn string) sqlmock.Sqlmock {
	_, mock, err := sqlmock.NewWithDSN(dsn)
	if err != nil {
		panic(err)
	}
	DB, err = gorm.Open("sqlmock", dsn)
	if err != nil {
		panic(err)
	}
	return mock
}
