package handlers

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"bytes"

	"database/sql"

	"fmt"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestWebfinger(t *testing.T) {
	//init
	config.InitBasicConfig(bytes.NewBuffer([]byte{}))
	config.Set("hostname", "peerpx.social")
	e := echo.New()

	// no resource
	req := httptest.NewRequest(echo.GET, "/.well-know/webfinger", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// bad syntax
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=accttoorop@peerpx.social,foobar", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// no acct
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=act:toorop@peerpx.social,foo:bar", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// not addr
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:tooroppeerpx.social", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// not local
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.org", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	// no such user
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.social", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	// DB failure
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.social", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(fmt.Errorf("boooooo!"))
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	// ok
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.social", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)

	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		//log.Print(rec.Body.String())
	}
}
