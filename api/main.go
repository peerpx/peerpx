package main

import (
	"os"

	"path"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/toorop/peerpx/api/controllers"
	"github.com/toorop/peerpx/api/middlewares"
	"github.com/toorop/peerpx/core"
	"github.com/toorop/peerpx/core/models"
)

func main() {
	var err error

	// load config
	if err = core.InitViper(); err != nil {
		log.Fatalf("unable to iny viper: %v ", err)
	}

	// init logger props
	// todo set formatter
	log.SetFormatter(&log.TextFormatter{})
	logDir := viper.GetString("log.dir")
	if logDir != "" {
		fd, err := os.OpenFile(path.Join(logDir, "peerpx.log"), os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatalf("unable to open log file: %v", err)
		}
		defer fd.Close()
		log.SetOutput(fd)
	} else {
		log.SetOutput(os.Stdout)
	}

	// init DB
	// todo mv to core
	db, err := gorm.Open("sqlite3", "peerpx.db")
	if err != nil {
		log.Fatalf("unable init DB: %v ", err)
	}
	defer db.Close()

	// Migrate the schema
	if err = db.AutoMigrate(&models.User{}, &models.Photo{}).Error; err != nil {
		log.Fatalf("unable to migrate DB: %v", err)
	}

	// init Echo

	e := echo.New()

	// routes

	// photo

	// upload
	e.POST("/api/v1/photo", controllers.Todo, middlewares.AuthRequired())

	// get photo
	e.GET("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// update photo properties
	e.PUT("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// delete photo
	e.DELETE("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// search
	e.GET("/api/v1/photo/search", controllers.PhotoSearch, middlewares.AuthRequired())

	log.Fatal(e.Start(":8080"))
}
