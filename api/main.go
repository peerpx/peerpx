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
)

func main() {
	var err error

	log.Info("On est ici")

	// load config
	if err = core.InitViper(); err != nil {
		log.Fatalf("unable to init viper: %v ", err)
	}

	// init logger props
	// todo set formatter & config
	log.SetFormatter(&log.TextFormatter{})
	logDir := viper.GetString("log.dir")
	if logDir != "" {
		var fd *os.File
		fd, err = os.OpenFile(path.Join(logDir, "peerpx.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Fatalf("unable to open log file: %v", err)
		}
		defer fd.Close()
		log.SetOutput(fd)
	} else {
		log.SetOutput(os.Stdout)
	}

	// init DB
	if err = core.InitDB(); err != nil {
		log.Fatalf("unable init DB: %v ", err)
	}
	defer core.DB.Close()
	log.Info("DB initialized")

	// Migrate the schema
	// TODO add option (its useless to migrate DB @each run)
	if err = core.DB.AutoMigrate(&models.User{}, &models.Photo{}).Error; err != nil {
		log.Fatalf("unable to migrate DB: %v", err)
	}

	// init datastore
	core.DS, err = core.NewDatastoreFs(viper.GetString("datastore.path"))
	if err != nil {
		log.Fatalf("unable to create datastore: %v", err)
	}

	// init Echo
	e := echo.New()

	// routes

	// photo

	// upload
	e.POST("/api/v1/photo", controllers.PhotoPost, middlewares.AuthRequired())

	// get photo
	// size:
	// 	max 	-> uploaded photo (modulo config max size)
	//  large	-> 2k ?
	// 	medium  -> 1k ?
	// 	small   -> 500
	//  usmall  -> 200
	//  int -> int
	e.GET("/api/v1/photo/:id:/:size", controllers.Todo, middlewares.AuthRequired())

	// get photo properties -> JSON object
	e.GET("/api/v1/photo/:id/properties", controllers.PhotoGetProperties)

	// update photo properties
	e.PUT("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// delete photo
	e.DELETE("/api/v1/photo/:id", controllers.Todo, middlewares.AuthRequired())

	// search
	e.GET("/api/v1/photo/search", controllers.Todo, middlewares.AuthRequired())

	e.Logger.Fatal(e.Start(":8080"))
}
