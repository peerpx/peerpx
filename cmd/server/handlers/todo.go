package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

// Todo for controllers to do
func Todo(c echo.Context) error {
	return c.String(http.StatusTeapot, "TODO")
}
