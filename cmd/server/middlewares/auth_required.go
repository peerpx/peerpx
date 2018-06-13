package middlewares

import (
	"database/sql"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/log"
)

// AuthRequired check auth
func AuthRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ac echo.Context) error {
			c := ac.(*AppContext)
			// cookie
			username, err := c.SessionGet("username")
			if err != nil {
				log.Errorf("%v - middleware.AuthRequired - unable to read session: %v", c.RealIP(), err)
				return echo.ErrCookieNotFound
			}
			if username != nil && username.(string) != "" {
				u, err := user.GetByUsername(username.(string))
				if err != nil {
					if err == sql.ErrNoRows {
						log.Errorf("%v - middleware.AuthRequired - cookie is present but user %s is not found", c.RealIP(), username.(string))
						return echo.ErrForbidden
					}
					log.Errorf("%v - middleware.AuthRequired - user.GetByUsername(%s) failed: %v", c.RealIP(), username.(string), err)
					return err
				}
				c.Set("u", u)
			} else {
				log.Infof("%v - auth required", c.RealIP())
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}
