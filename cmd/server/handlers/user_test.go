package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
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
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
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
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
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

	// marshall(user) failed

	// OK
	db.Mock.ExpectPrepare("^INSERT INTO users (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	data = `{"Email": "bar@foo.com", "Username": "john", "Password": "dhfsdjhfjk"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user", strings.NewReader(data))
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
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
	e := echo.New()
	// bad data
	req := httptest.NewRequest(echo.POST, "/api/v1/user/login", errReader(0))
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))

	if assert.NoError(t, UserLogin(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Nil(t, response.Data)
			assert.False(t, response.Success)
			assert.Equal(t, "requestBodyNotReadable", response.Code)
		}
	}

	// bad body (not json)
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", nil)
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, UserLogin(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Nil(t, response.Data)
			assert.False(t, response.Success)
			assert.Equal(t, "requestBodyNotValidJson", response.Code)
		}
	}

	// no such user
	body := `{"login":"john", "password":"secret"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)

	if assert.NoError(t, UserLogin(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.Equal(t, "noSuchUser", response.Code)
		}
	}

	// error
	body = `{"login":"john", "password":"secret"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))

	if assert.NoError(t, UserLogin(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.Equal(t, "userLoginFailed", response.Code)
		}
	}

	// ok
	body = `{"login":"john", "password":"secret"}`
	req = httptest.NewRequest(echo.POST, "/api/v1/user/login", strings.NewReader(body))
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)

	if assert.NoError(t, UserLogin(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
			if assert.NotNil(t, response.Data) {
				u := new(user.User)
				if assert.NoError(t, json.Unmarshal(response.Data, u)) {
					assert.Equal(t, uint(1), u.ID)
					assert.Equal(t, "john", u.Username)
					assert.Equal(t, "john@doe.com", u.Email)
					username, _ := c.SessionGet("username")
					assert.Equal(t, "john", username)
				}
			}
		}
	}
}

func TestUserMe(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/api/v1/user/me", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))

	// user not authenticated (should not happen)
	if assert.NoError(t, UserMe(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, response.Code, "userNotInContext")
		}
	}

	// ok
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	u := new(user.User)
	u.ID = 1
	u.Email = "foo@bar.com"
	c.Set("u", *u)
	if assert.NoError(t, UserMe(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
			u = new(user.User)
			if assert.NoError(t, json.Unmarshal(response.Data, u)) {
				assert.Equal(t, uint(1), u.ID)
				assert.Equal(t, "foo@bar.com", u.Email)
			}
		}
	}
}
