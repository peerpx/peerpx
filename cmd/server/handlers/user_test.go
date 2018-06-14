package handlers

import (
	"net/http/httptest"
	"testing"

	"encoding/json"

	"net/http"

	"strings"

	"errors"

	"database/sql"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestUserCreate(t *testing.T) {
	e := echo.New()
	config.Set("usernameMaxLength", "5")
	config.Set("usernameMinLength", "3")

	// read body failed
	req := httptest.NewRequest(echo.POST, "/api/v1/user", errReader(0))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.Equal(t, "requestBodyNotReadable", response.Code)
		}
	}

	// bad body (not json)
	req = httptest.NewRequest(echo.POST, "/api/v1/user", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.Equal(t, "requestBodyNotValidJson", response.Code)
		}
	}

	// create failed (bad input not an email)
	data := `{"Email": "barfoo.com", "Username": "john", "Password": "dhfsdjhfjk"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user", strings.NewReader(data))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.Equal(t, "userCreateFailed", response.Code)
			assert.True(t, strings.HasSuffix(response.Message, "barfoo.com is not a valid email"))
		}
	}

	// OK
	db.Mock.ExpectPrepare("^INSERT INTO users (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	data = `{"Email": "bar@foo.com", "Username": "john", "Password": "dhfsdjhfjk"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user", strings.NewReader(data))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
			assert.NotNil(t, response.Data)
			u := new(user.User)
			if assert.NoError(t, json.Unmarshal(response.Data, u)) {
				assert.Equal(t, uint(1), u.ID)
			}
		}
	}
}

func TestUserLogin(t *testing.T) {
	// bad data
	// bad body (not json)
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/api/v1/user/login", nil)
	rec := httptest.NewRecorder()
	c := &middlewares.AppContext{e.NewContext(req, rec), sessions.NewCookieStore([]byte("cookieAuthKey"), []byte("cookieEncryptionKey"))}

	if assert.NoError(t, UserLogin(c)) {
		response := new(userLoginResponse)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.Nil(t, response.User)
			assert.Equal(t, "bad json", response.Msg)
		}
	}

	// no such user
	body := `{"login":"john", "password":"secret"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = &middlewares.AppContext{e.NewContext(req, rec), sessions.NewCookieStore([]byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"), []byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"))}
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)
	if assert.NoError(t, UserLogin(c)) {
		response := new(userLoginResponse)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.Nil(t, response.User)
			assert.Equal(t, "no such user", response.Msg)
		}
	}

	// internal server error
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = &middlewares.AppContext{e.NewContext(req, rec), sessions.NewCookieStore([]byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"), []byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"))}
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))
	if assert.NoError(t, UserLogin(c)) {
		response := new(userLoginResponse)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.Nil(t, response.User)
			assert.Equal(t, "unable to login", response.Msg)
		}
	}

	// created
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = &middlewares.AppContext{e.NewContext(req, rec), sessions.NewCookieStore([]byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"), []byte("xN4vP672vbvtb7cp7HuTH4XzD8HZbLV4"))}
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	if assert.NoError(t, UserLogin(c)) {
		response := new(userLoginResponse)
		assert.Equal(t, http.StatusOK, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.NotNil(t, response.User)
			assert.Equal(t, "john", response.User.Username)
			assert.Equal(t, "john@doe.com", response.User.Email)
			assert.Equal(t, "", response.Msg)
		}
	}

}
