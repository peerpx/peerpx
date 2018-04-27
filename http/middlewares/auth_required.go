package middlewares

import (
	"github.com/labstack/echo"
)

// AuthRequired check auth
func AuthRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// get header x-api-key
			key := c.Request().Header.Get("x-api-key")
			if key != "QmU2TQthpXDj8QNK6jyqpWsjgDmr3E9Hn3F6zTahGGvZUC" {
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}
