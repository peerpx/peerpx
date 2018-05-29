package main

import (
	"os/user"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/peerpx/peerpx/cmd/server/handlers"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
	"github.com/peerpx/peerpx/services/log"
	"github.com/spf13/viper"
	"github.com/toorop/peerpx/api/controllers"
	"github.com/toorop/peerpx/api/middlewares"

	"os"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	var err error

	// init logger
	log.InitBasicLogger(os.Stdout)

	// load config
	if err = config.InitViper(); err != nil {
		log.Errorf("viper initialization failed : %v ", err)
		os.Exit(1)
	}

	// init logger props

	// init DB
	if err = db.InitDB(); err != nil {
		log.Errorf("DB initialization failed: %v ", err)
		os.Exit(1)
	}
	defer db.DB.Close()
	//log.Info("DB initialized")

	// Migrate the schema
	// TODO add option (its useless to migrate DB @each run)
	if err = db.DB.AutoMigrate(&user.User{}, &photo.Photo{}).Error; err != nil {
		log.Errorf("unable to migrate DB: %v", err)
		os.Exit(1)
	}

	// init datastore
	if err = datastore.InitFilesystemDatastore(viper.GetString("datastore.path")); err != nil {
		log.Errorf("datastore initialization failed: %v", err)
		os.Exit(1)
	}

	// init
	e := echo.New()

	// add CORS
	e.Use(middleware.CORS())

	// routes

	// photo

	// upload
	e.POST("/api/v1/photo", handlers.PhotoPost, middlewares.AuthRequired())

	// get photo
	// size:
	// 	max 	-> uploaded photo (modulo config max size)
	//  xl -> 2k ?
	// 	l  -> 1k ?
	// 	m  -> 500
	//  s  -> 200
	//  xs ->
	e.GET("/api/v1/photo/:id/:size", controllers.PhotoGet)

	// resize photo by height (in pixel)
	e.GET("/api/v1/photo/:id/height/:height", controllers.PhotoResize)

	// returns photo resized by width
	e.GET("/api/v1/photo/:id/width/:width", controllers.PhotoResize)

	// get photo properties -> JSON object
	e.GET("/api/v1/photo/:id/properties", controllers.PhotoGetProperties)

	// update photo properties
	e.PUT("/api/v1/photo", handlers.PhotoPut, middlewares.AuthRequired())

	// delete photo
	e.DELETE("/api/v1/photo/:id", controllers.PhotoDel, middlewares.AuthRequired())

	// search
	e.GET("/api/v1/photo/search", controllers.PhotoSearch)

	//user

	// add user
	e.POST("/api/v1/user", controllers.UserPost)

	e.Logger.Fatal(e.Start(":8080"))

}
