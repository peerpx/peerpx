package main

import (
	"os"
	"path"
	"path/filepath"

	"net/http"

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
	"github.com/peerpx/peerpx/services/log"
)

func main() {
	var err error

	// init logger
	log.InitBasicLogger(os.Stdout)

	// get working dir
	workingDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Errorf("get working dir failed: %v", err)
		os.Exit(1)
	}

	// load config
	if err = config.InitBasicConfigFromFile(path.Join(workingDir, "peerpx.conf")); err != nil {
		log.Errorf("init config failed : %v ", err)
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
	if err = datastore.InitFilesystemDatastore(config.GetStringDefault(("datastore.path"), path.Join(workingDir, "datastore"))); err != nil {
		log.Errorf("datastore initialization  PLOPfailed: %v", err)
		os.Exit(1)
	}

	// init
	e := echo.New()

	// add custom context
	e.Use(middlewares.Context)

	// add CORS
	if !config.GetBoolDefault("prod", true) {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000/", "*"},
			AllowCredentials: true,
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-Api-Key"},
			AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		}))
	}

	// routes

	////
	// user

	// create user
	e.POST("/api/v1/user", handlers.UserCreate)

	// get me
	e.GET("/api/v1/user/me", handlers.UserMe, middlewares.AuthRequired())
	// TODO remove
	e.POST("/user/me", handlers.UserMe, middlewares.AuthRequired())

	// update user
	e.PUT("/api/v1/user", handlers.Todo)

	// delete user
	e.DELETE("/api/v1/user", handlers.Todo)

	// login
	e.POST("/api/v1/user/login", handlers.UserLogin)

	// logout
	e.POST("/api/v1/user/logout", handlers.Todo)

	// check if pseudo is available
	e.GET("/api/v1/user/pseudo/:pseudo/is-available", handlers.Todo)

	////
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
	e.GET("/api/v1/photo/:id/:size", handlers.PhotoGet)

	// resize photo by height (in pixel)
	e.GET("/api/v1/photo/:id/height/:height", handlers.PhotoResize)

	// returns photo resized by width
	e.GET("/api/v1/photo/:id/width/:width", handlers.PhotoResize)

	// get photo properties -> JSON object
	e.GET("/api/v1/photo/:id/properties", handlers.PhotoGetProperties)

	// update photo properties
	e.PUT("/api/v1/photo", handlers.PhotoPut, middlewares.AuthRequired())

	// delete photo
	e.DELETE("/api/v1/photo/:id", handlers.PhotoDel, middlewares.AuthRequired())

	// search
	e.GET("/api/v1/photo/search", handlers.PhotoSearch)

	// API 404
	e.Any("/api/*", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})

	/////
	// Client
	e.Static("/", "./www")
	e.File("/", "./www/index.html")
	e.File("/signup", "./www/index.html")
	e.Logger.Fatal(e.Start(":8080"))

}
