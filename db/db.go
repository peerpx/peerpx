package db

import "github.com/jinzhu/gorm"

// DB is the database connector
var DB *gorm.DB

// InitDB initialize database connector
// todo: get options/type from viper
func InitDB() (err error) {
	DB, err = gorm.Open("sqlite3", "peerpx.db")
	return err
}
