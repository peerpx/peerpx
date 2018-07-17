package context

import (
	"fmt"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/log"
	"github.com/satori/go.uuid"
)

// AppContext extends echo.Context
// add session management via encrypted cookie
type AppContext struct {
	echo.Context
	CookieStore *sessions.CookieStore
	UUID        string
}

func NewMockedContext(c echo.Context) *AppContext {
	return &AppContext{
		c,
		sessions.NewCookieStore([]byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"), []byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4")),
		uuid.Must(uuid.NewV4()).String(),
	}
}

// SetCookieStore CookieStore setter
func (c *AppContext) SetCookieStore(cs *sessions.CookieStore) {
	c.CookieStore = cs
}

// GetCookieStore CookieStore getter
func (c *AppContext) GetCookieStore() *sessions.CookieStore {
	return c.CookieStore
}

// SessionGet get data from session
func (c *AppContext) SessionGet(key string) (interface{}, error) {
	session, err := c.CookieStore.Get(c.Request(), "ppx")
	if err != nil {
		return nil, err
	}
	return session.Values[key], nil
}

// SessionSet set data in session
func (c *AppContext) SessionSet(key string, value interface{}) error {
	session, err := c.CookieStore.Get(c.Request(), "ppx")
	if err != nil {
		return err
	}
	session.Values[key] = value
	return session.Save(c.Request(), c.Response().Writer)
}

// SessionExpire expire the current session
func (c *AppContext) SessionExpire() error {
	session, err := c.CookieStore.Get(c.Request(), "ppx")
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(c.Request(), c.Response())
}

// GetWantedContentType returns wanted content type
// Warning file extension >> header accept (yes i know...)
// Warning supported CT html, json, atom
func (c *AppContext) GetWantedContentType() string {
	uri := strings.ToLower(c.Request().RequestURI)
	if strings.HasSuffix(uri, ".json") {
		return "json"
	} else if strings.HasSuffix(uri, ".atom") {
		return "atom"
	} else {
		accept := strings.ToLower(c.Request().Header.Get("accept"))
		if strings.HasPrefix(accept, "application/json") {
			return "json"
		}
		if strings.HasPrefix(accept, "application/atom+xml") {
			return "atom"
		}
	}
	return ""
}

// LogInfo is the info level logger
func (c *AppContext) LogInfo(v ...interface{}) {
	log.Info(fmt.Sprintf("%s - %s - ", c.RealIP(), c.UUID), fmt.Sprintln(v...))
}

// LogInfof is the info level logger
func (c *AppContext) LogInfof(format string, v ...interface{}) {
	log.Infof(fmt.Sprintf("%s - %s - %s ", c.RealIP(), c.UUID, format), v...)
}

// LogError is the error level logger
func (c *AppContext) LogError(v ...interface{}) {
	log.Error(fmt.Sprintf("%s - %s - ", c.RealIP(), c.UUID), fmt.Sprintln(v...))
}

// LogErrorf is the info level logger
func (c *AppContext) LogErrorf(format string, v ...interface{}) {
	log.Errorf(fmt.Sprintf("%s - %s - %s ", c.RealIP(), c.UUID, format), v...)
}

// Context app context
func Context(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &AppContext{
			c,
			sessions.NewCookieStore([]byte(config.GetStringP("cookieAuthKey")), []byte(config.GetStringP("cookieEncrytionKey"))),
			uuid.Must(uuid.NewV4()).String(),
		}
		return h(cc)
	}
}
