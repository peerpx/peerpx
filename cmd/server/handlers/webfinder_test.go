package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"github.com/peerpx/peerpx/services/log"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const userPubKey = `-----BEGIN PUBLIC KEY-----
MIIBCgKCAQEAqVIsY/YRF/+Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV+LpxZesn
+jspCUXEID0bowbUXly+QkBsA3ZBFOAE4vmd+XQ3ukt+aHHWJnJVpZjrMScDIYrJ
RENXAMyW4yZ1tnL66efm5/qsYypqOEICLr27A0+yIwlJ4vjlziy+rEwFihdJKorv
RBCAiYBUgio7l9Y+Oo0kqd/ZL8DtBHYqsSyTcRcHL/s/O2Ktyxo7cUsvelmTClS2
zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp+yF4N/JtXGZ
+oP/W8u+1TQl1G54j0MFyalZjtEzEpe+RQIDAQAB
-----END PUBLIC KEY-----`

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
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(fmt.Errorf("boooooo"))
	if assert.NoError(t, Webfinger(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	// ok
	req = httptest.NewRequest(echo.GET, "/.well-know/webfinger?resource=acct:toorop@peerpx.social", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	row := sqlmock.NewRows([]string{"id", "username", "email", "public_key"}).AddRow(1, "john", "john@doe.com", userPubKey)
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)

	if assert.NoError(t, Webfinger(c)) {

		assert.Equal(t, http.StatusOK, rec.Code)

		log.Info(rec.Body.String())
	}
}
