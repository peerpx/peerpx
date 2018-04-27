package main

import (
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"github.com/toorop/peerpx/core/models"
)

func main() {
	var err error

	// load config
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("unable to read config: %v ", err)
	}

	// init DB
	db, err := gorm.Open("sqlite3", "peerpx.db")
	if err != nil {
		log.Fatalf("unable init DB: %v ", err)
	}
	defer db.Close()

	// Migrate the schema
	if err = db.AutoMigrate(&models.User{}, &models.Photo{}).Error; err != nil {
		log.Fatalf("unable to migrate DB: %v", err)
	}

	// init app logger TODO

	// init Echo

	e := echo.New()

	// routes
	e.GET("/", func(c echo.Context) error {
		log.Println("toto")
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":8080"))
}
