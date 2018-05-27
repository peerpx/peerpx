package main

import (
	"os"

	"path"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/peerpx/peerpx/cmd/server/handlers"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/toorop/peerpx/api/controllers"
)

func main() {
	var err error

	// load config
	if err = config.InitViper(); err != nil {
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
	if err = db.InitDB(); err != nil {
		log.Fatalf("unable init DB: %v ", err)
	}
	defer db.DB.Close()
	log.Info("DB initialized")

	// Migrate the schema
	// TODO add option (its useless to migrate DB @each run)
	if err = db.DB.AutoMigrate(&user.User{}, &photo.Photo{}).Error; err != nil {
		log.Fatalf("unable to migrate DB: %v", err)
	}

	// init datastore
	if err = datastore.InitFilesystemDatastore(viper.GetString("datastore.path")); err != nil {
		log.Fatalf("unable to create datastore: %v", err)
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

	e.Logger.Fatal(e.Start(":8080"))
}
