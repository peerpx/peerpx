package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"text/template"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/context"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/pkg/cryptobox"
	"github.com/peerpx/peerpx/services/config"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/json"
)

// WebfingerAcct basic implementation
func WebfingerAcct(ac echo.Context) error {
	c := ac.(*context.AppContext)
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

	// generate response body
	t, err := tplBox.MustString("webfinger/acct.json")
	if err != nil {
		c.LogErrorf("handlers.WebfingerAcct - tplBox.MustString(webfinger/acct.json) failed: %v", err)
		return c.String(http.StatusInternalServerError, "internal server error")
	}
	tpl, err := template.New("acct").Parse(t)
	if err != nil {
		c.LogErrorf("handlers.WebfingerAcct - template new failed: %v", err)
		return c.String(http.StatusInternalServerError, "internal server error")
	}

	tplData := struct {
		HostName string
		UserName string
		MagicKey string
	}{
		HostName: config.GetString("hostname"),
		UserName: u.Username,
		MagicKey: magicKey,
	}
	b := bytes.NewBuffer(nil)
	if err = tpl.Execute(b, tplData); err != nil {
		c.LogErrorf("handlers.WebfingerAcct - template execute failed: %v", err)
		return c.String(http.StatusInternalServerError, "internal server error")
	}

	m := minify.New()
	m.AddFunc("application/json", json.Minify)
	out, err := m.Bytes("application/json", b.Bytes())
	if err != nil {
		// TODO log err
		return c.String(http.StatusInternalServerError, "i'm sorry dave, something went wrong")
	}
	return c.Blob(200, "application/jrd+json; charset=utf-8", out)
}
