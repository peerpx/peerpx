package middlewares

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
)

// AppContext extends echo.Context
// add session management via encrypted cookie
type AppContext struct {
	echo.Context
	CookieStore *sessions.CookieStore
}

// SetCookieStore CookieStore setter
func (a *AppContext) SetCookieStore(cs *sessions.CookieStore) {
	a.CookieStore = cs
}

// GetCookieStore CookieStore getter
func (a *AppContext) GetCookieStore() *sessions.CookieStore {
	return a.CookieStore
}

// SessionGet get data from session
func (a *AppContext) SessionGet(key string) (interface{}, error) {
	session, err := a.CookieStore.Get(a.Request(), "ppx")
	if err != nil {
		return nil, err
	}
	return session.Values[key], nil
}

// SessionSet set data in session
func (a *AppContext) SessionSet(key string, value interface{}) error {
	session, err := a.CookieStore.Get(a.Request(), "ppx")
	if err != nil {
		return err
	}
	session.Values[key] = value
	return session.Save(a.Request(), a.Response().Writer)
}

func Context(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &AppContext{c, nil}
		cc.SetCookieStore(sessions.NewCookieStore([]byte(config.GetStringP("cookieAuthKey")), []byte(config.GetStringP("cookieEncrytionKey"))))
		return h(cc)
	}
}