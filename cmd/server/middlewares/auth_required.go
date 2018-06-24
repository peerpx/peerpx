package middlewares

import (
	"database/sql"

	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/context"
	"github.com/peerpx/peerpx/cmd/server/handlers"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/log"
)

// AuthRequired check auth
func AuthRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ac echo.Context) error {
			c := ac.(*context.AppContext)
			response := handlers.NewApiResponse(c.UUID)

			// cookie
			username, err := c.SessionGet("username")
			if err != nil {
				log.Errorf("%s - %s - middleware.AuthRequired - unable to read session: %v", c.RealIP(), c.UUID, err)
				return echo.ErrCookieNotFound
			}
			if username != nil && username.(string) != "" {
				u, err := user.GetByUsername(username.(string))
				if err != nil {
					if err == sql.ErrNoRows {
						c.LogInfof("middleware.AuthRequired - cookie is present but user %s is not found", username.(string))
						// expire session
						if err = c.SessionExpire(); err != nil {
							msg := fmt.Sprintf("%s - %s - middleware.AuthRequired -  sessionExpire failed: %v", c.RealIP(), response.UUID, err)
							return response.Error(c, http.StatusInternalServerError, "userMarshalFailed", msg)
						}
						return response.OK(c, http.StatusForbidden)
					}
					log.Errorf("%s - %s - middleware.AuthRequired - user.GetByUsername(%s) failed: %v", c.RealIP(), c.UUID, username.(string), err)
					return err
				}
				c.Set("u", u)
			} else {
				log.Infof("%s - auth required", c.RealIP())
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}
