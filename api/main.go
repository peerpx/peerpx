package main

import (
	"os"

	"path"

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
	"github.com/toorop/peerpx/db"
)

func main() {
	var err error

	// load config
	if err = core.InitViper(); err != nil {
		log.Fatalf("unable to iny viper: %v ", err)
	}

	// init logger props
	// todo set formatter & config
	log.SetFormatter(&log.TextFormatter{})
	logDir := viper.GetString("log.dir")
	if logDir != "" {
		var fd *os.File
		fd, err = os.OpenFile(path.Join(logDir, "peerpx.log"), os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatalf("unable to open log file: %v", err)
		}
		defer fd.Close()
		log.SetOutput(fd)
	} else {
		log.SetOutput(os.Stdout)
	}

	// init DB
	if err = db.InitDB(); err != nil {
		log.Fatalf("unable init DB: %v ", err)
	}
	defer db.DB.Close()

	// Migrate the schema
	// TODO add option (its useless to migrate DB @each run)
	if err = db.DB.AutoMigrate(&models.User{}, &models.Photo{}).Error; err != nil {
		log.Fatalf("unable to migrate DB: %v", err)
	}

	// init datastore

	// init Echo
	e := echo.New()

	// routes

	// photo

	// upload
	e.POST("/api/v1/photo", controllers.Todo, middlewares.AuthRequired())

	// get photo -> RAW photo
	e.GET("/api/v1/photo/:id:/:size", controllers.Todo, middlewares.AuthRequired())

	// get photo properties -> JSON object
	e.GET("/api/v1/photo/:id:/properties", controllers.Todo, middlewares.AuthRequired())

	// update photo properties
	e.PUT("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// delete photo
	e.DELETE("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// search
	e.GET("/api/v1/photo/search", controllers.PhotoSearch, middlewares.AuthRequired())

	log.Fatal(e.Start(":8080"))
}
