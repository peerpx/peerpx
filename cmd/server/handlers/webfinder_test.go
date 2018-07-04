package handlers

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"bytes"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestWebfinger(t *testing.T) {
	config.InitBasicConfig(bytes.NewBuffer([]byte{}))
	config.Set("hostname", "peerpx.social")
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.social", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)

	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		//log.Print(rec.Body.String())
	}
}
