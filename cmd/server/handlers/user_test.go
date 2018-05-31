package handlers

import (
	"net/http/httptest"
	"testing"

	"encoding/json"

	"net/http"

	"strings"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestUserCreate(t *testing.T) {

	viper.Set("usernameMaxLength", 5)
	viper.Set("usernameMinLength", 3)

	// bad body (not json)
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/api/v1/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		response := new(userCreateResponse)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.Nil(t, response.User)
			assert.Equal(t, "bad json", response.Msg)
		}
	}

	// bad input (not an email)
	data := `{"Email": "barfoo.com", "Username": "john", "Password": "dhfsdjhfjk"}`

	req = httptest.NewRequest(echo.POST, "/api/v1/user", strings.NewReader(data))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		response := new(userCreateResponse)
		assert.Equal(t, http.StatusCreated, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.Nil(t, response.User)
			assert.Equal(t, "barfoo.com is not a valid email", response.Msg)
		}
	}

	// OK
	// mocked DB
	mock := db.InitMockedDB("sqlmock_db_ctrlusercreate")
	defer db.DB.Close()
	mock.ExpectExec("^INSERT INTO \"users\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))
	data = `{"Email": "bar@foo.com", "Username": "john", "Password": "dhfsdjhfjk"}`

	req = httptest.NewRequest(echo.POST, "/api/v1/user", strings.NewReader(data))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, UserCreate(c)) {
		response := new(userCreateResponse)
		assert.Equal(t, http.StatusCreated, rec.Code)
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), response)) {
			assert.NotNil(t, response.User)
			assert.Equal(t, "", response.Msg)
		}
	}
}
