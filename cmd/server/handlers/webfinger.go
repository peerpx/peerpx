package handlers

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"database/sql"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/pkg/cryptobox"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/log"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/json"
)

const webFingerResponseTpl = `{
	"subject": "acct:%s",
	"links": [
		{
			"rel": "self",
			"type": "application/activity+json",
			"href": "https://%s/%s"
		},
		{
			"rel": "magic-public-key",
			"href": "data:application/magic-public-key,%s"
},
	]
}`

// Webfinger basic implementation (POC)
func Webfinger(c echo.Context) error {
	// get resource
	resource := c.QueryParam("resource")
	if resource == "" {
		return c.String(http.StatusBadRequest, "missing resource parameter")
	}
	resource = strings.ToLower(resource)
	p := strings.SplitAfterN(resource, ":", 2)
	if len(p) != 2 {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// p[0] must be acct
	if p[0] != "acct:" {
		return c.String(http.StatusBadRequest, fmt.Sprintf("%s is not valid", p[0]))
	}

	// valid ? Same syntax as mail so...
	if _, err := mail.ParseAddress(p[1]); err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("%s is not a valid actor id", p[1]))
	}

	// local domain ?
	usernameDomain := strings.Split(p[1], "@")
	if usernameDomain[1] != config.GetString("hostname") {
		return c.String(http.StatusNotFound, fmt.Sprintf("%s is not referenced here", p[1]))
	}

	// Get user
	u, err := user.GetByUsername(usernameDomain[0])
	if err != nil {
		if err == sql.ErrNoRows {
			return c.String(http.StatusNotFound, "not found ")
		}
		return c.String(http.StatusInternalServerError, "i'm sorry dave, something went wrong")
	}

	// magicKey
	magicKey, err := cryptobox.RSAGetMagicKey(u.PublicKey.String)
	if err != nil {
		return c.String(http.StatusInternalServerError, "get magic key failed")
	}

	m := minify.New()
	m.AddFunc("application/json", json.Minify)
	out, err := m.String("application/json", fmt.Sprintf(webFingerResponseTpl, p[1], config.GetString("hostname"), usernameDomain[0], magicKey))
	if err != nil {
		log.Errorf("Parse failed %v", err)
		return c.String(http.StatusInternalServerError, "i'm sorry dave, something went wrong")
	}
	return c.String(http.StatusOK, out)
}
