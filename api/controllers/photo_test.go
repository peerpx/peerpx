package controllers

import (
	"io/ioutil"
	"testing"

	"bytes"
	"net/http/httptest"

	"net/http"

	"encoding/json"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/peerpx/peerpx/core"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"mime/multipart"
	"os"
	"io"
	"github.com/jinzhu/gorm"
)


func handleErr(err error){
	if err != nil {
		panic(err)
	}
}


func TestPhotoPost(t *testing.T) {

	// init viper (small values -> photo will be re-encoded)
	viper.Set("photo.maxWidth", 100)
	viper.Set("photo.maxHeight", 100)

	//  init mocked datastore
	core.DS = core.NewDatastoreMocked([]byte{}, nil)

	// mocked DB
	mock := core.InitMockedDB("sqlmock_db_0")
	defer core.DB.Close()
	mock.ExpectExec("^INSERT INTO \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))


	data := `{"Name":"ma super photo", "Description":" ma description"}`

	body := new(bytes.Buffer)
	writer :=multipart.NewWriter(body)

	handleErr(writer.WriteField("data", data))

	file, err := os.Open("../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()


	part, err := writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)

	_, err  = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())

/*
	bodyByte, err := ioutil.ReadAll(body)
	handleErr(err)
	println(string(bodyByte))
*/

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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

/*
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
*/

func TestPhotoGetPropertiesByHash(t *testing.T) {
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

func TestPhotoGetPropertiesByHashNotFound(t *testing.T) {
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

func TestPhotoGet(t *testing.T) {
	//  init mocked datastore
	core.DS = core.NewDatastoreMocked([]byte{1, 2, 3}, nil)

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", "mocked")
	c.Set("size", "small")
	if assert.NoError(t, PhotoGet(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		data, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2, 3}, data)
	}
}

func TestPhotoGetNotfound(t *testing.T) {
	//  init mocked datastore
	core.DS = core.NewDatastoreMocked(nil, core.ErrNotFoundInDatastore)

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", "mocked")
	c.Set("size", "small")
	if assert.NoError(t, PhotoGet(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestPhotoResize(t *testing.T) {
}

func TestPhotoDel(t *testing.T) {
	mock := core.InitMockedDB("sqlmock_db_3")
	mock.ExpectExec("DELETE FROM \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))
	defer core.DB.Close()
	core.DS = core.NewDatastoreMocked(nil, nil)
	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("id", "mocked")
	if assert.NoError(t, PhotoDel(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestPhotoSearchNoArgs(t *testing.T) {
	// mocked DB
	mock := core.InitMockedDB("sqlmock_db_4")
	defer core.DB.Close()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	//c.Set("id", "mocked")

	// test "valid" hash
	rows := sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(rows)
	if assert.NoError(t, PhotoSearch(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
