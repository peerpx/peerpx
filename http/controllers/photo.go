package controllers

import (
	"net/http"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

// PhotoSearch API ctrl for searching photos
func PhotoSearch(c echo.Context) error {
	log.Print("toto")
	log.Error("toto")
	return c.String(http.StatusOK, "TODO")
}
