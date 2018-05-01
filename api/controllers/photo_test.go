package controllers

import (
	"io/ioutil"
	"testing"

	"bytes"
	"net/http/httptest"

	"net/http"

	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/toorop/peerpx/core"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPhotoPost(t *testing.T) {
	// init viper (small values -> photo will be re-encoded)
	viper.Set("photo.maxWidth", 100)
	viper.Set("photo.maxHeight", 100)

	//  init mocked datastore
	core.DS = core.NewDatastoreMocked()

	// mocked DB
	mock := core.InitMockedDB("sqlmock_db_0")
	defer core.DB.Close()
	mock.ExpectExec("^INSERT INTO \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))

	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", bytes.NewBuffer(photoBytes))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, PhotoPost(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		var response PhotoPostResponse
		err = json.Unmarshal(resp, &response)
		assert.NoError(t, err)
		assert.Equal(t, "H62MqsYPjtrQ56bgEJyaMVSGNJH3koXkBHgpj4uigR8T", response.PhotoID)
	}

}

func TestPhotoPostNotAPhoto(t *testing.T) {
	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/not-a-photo.jpg")
	if err != nil {
		panic(err)
	}
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", bytes.NewBuffer(photoBytes))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, PhotoPost(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		var response PhotoPostResponse
		err = json.Unmarshal(resp, &response)
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), response.Code)
	}
}

func TestPhotoGetByHash(t *testing.T) {
	// mocked DB
	mock := core.InitMockedDB("sqlmock_db_1")
	defer core.DB.Close()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", "mocked")

	// test "valid" hash
	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "mocked")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(rows)
	if assert.NoError(t, PhotoGetProperties(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestPhotoGetByHashNotFound(t *testing.T) {
	// mocked DB
	mock := core.InitMockedDB("sqlmock_db_2")
	defer core.DB.Close()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", "mocked")

	mock.ExpectQuery("^SELECT(.*)").WillReturnError(gorm.ErrRecordNotFound)
	if assert.NoError(t, PhotoGetProperties(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
