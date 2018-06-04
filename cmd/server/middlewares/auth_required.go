package middlewares

import (
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/entities/user"
	log "github.com/sirupsen/logrus"
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
					if err == user.ErrNoSuchUser {
						log.Errorf("%v - middleware.AuthRequired - ucookie is present but user %s is not found", c.RealIP(), username.(string))
						return echo.ErrUnauthorized
					}
					log.Errorf("%v - middleware.AuthRequired - user.GetByUsername(%s) failed: %v", c.RealIP(), username.(string), err)
					return err
				}
				c.Set("u", u)
			} else {
				// TODO refactor
				// get header x-api-key
				//key := c.Request().Header.Get("x-api-key")
				//if key != "QmU2TQthpXDj8QNK6jyqpWsjgDmr3E9Hn3F6zTahGGvZUC" {
				log.Infof("%v - auth required", c.RealIP())
				return echo.ErrForbidden
				//}
			}
			return next(c)
		}
	}
}
