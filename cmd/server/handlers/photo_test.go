package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"net/http"

	"io"
	"os"

	"encoding/json"

	"errors"

	"database/sql"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func init() {
	db.InitMockedDatabase()
}

func TestPhotoCreate(t *testing.T) {
	e := echo.New()
	config.InitBasicConfig(strings.NewReader(""))
	config.Set("photo.maxWidth", "100")
	config.Set("photo.maxHeight", "100")
	config.Set("hostname", "localhost")
	//  init mocked datastore

	// request is not multipart
	req := httptest.NewRequest(echo.POST, "/api/v1/photo", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "reqNotMultipart", response.Code)
		}
	}

	// no properties in multipart
	foo := `{"foo":"bar"}`
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	handleErr(writer.WriteField("foo", foo))
	handleErr(writer.Close())

	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "reqBadMultipartProperties", response.Code)
		}
	}

	// properties bad format
	properties := "foobar"
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	handleErr(writer.Close())

	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "reqBadPhotoProperties", response.Code)
		}
	}

	// no file
	properties = `{"name":"ma super photo", "description":" ma description"}`
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "reqBadMultipartFile", response.Code)
		}
	}

	// bad mime type
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	file, err := os.Open("../../../etc/samples/photos/not-a-photo.jpg")
	handleErr(err)
	defer file.Close()
	part, err := writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)
	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "unsupportedPhotoFormat", response.Code)
		}
	}

	// datastore failed
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	file, err = os.Open("../../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()
	part, err = writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)
	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	datastore.InitMokedDatastore([]byte{}, errors.New("mocked"))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "datastoreFailed", response.Code)

		}
	}

	// db failed
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	file, err = os.Open("../../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()
	part, err = writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)
	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))

	datastore.InitMokedDatastore([]byte{}, nil)
	//DB
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		ExpectExec().
		WillReturnError(errors.New("mocked"))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "dbCreateFailed", response.Code)
		}
	}

	// duplicate
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	file, err = os.Open("../../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()
	part, err = writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)
	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))

	datastore.InitMokedDatastore([]byte{}, nil)
	//DB
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		ExpectExec().
		WillReturnError(errors.New("UNIQUE CONSTRAINT blabla"))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusConflict, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "duplicate", response.Code)
		}
	}

	// ok
	body = new(bytes.Buffer)
	writer = multipart.NewWriter(body)
	handleErr(writer.WriteField("properties", properties))
	file, err = os.Open("../../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()
	part, err = writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)
	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())
	req = httptest.NewRequest(echo.POST, "/api/v1/photo", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))

	datastore.InitMokedDatastore([]byte{}, nil)
	//DB
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	if assert.NoError(t, PhotoCreate(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
			// unmashall photo
			p := new(photo.Photo)
			err = json.Unmarshal(response.Data, p)
			if assert.NoError(t, err) {
				assert.Equal(t, uint(1), p.ID)
				assert.Equal(t, "H62MqsYPjtrQ56bgEJyaMVSGNJH3koXkBHgpj4uigR8T", p.Hash)
			}
		}
	}
}

func TestPhotoGetProperties(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/api/v1/photo", nil)

	// db failed
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))
	if assert.NoError(t, PhotoGetProperties(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "getByHashFailed", response.Code)
		}
	}

	// not found
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)
	if assert.NoError(t, PhotoGetProperties(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.False(t, response.Success)
			assert.Equal(t, "notFound", response.Code)
		}
	}

	// Ok
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	row := sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	if assert.NoError(t, PhotoGetProperties(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
			p := new(photo.Photo)
			err = json.Unmarshal(response.Data, p)
			if assert.NoError(t, err) {
				assert.Equal(t, uint(1), p.ID)
				assert.Equal(t, "mocked", p.Hash)
			}
		}
	}
}

func TestPhotoGet(t *testing.T) {
	e := echo.New()
	// not found
	datastore.InitMokedDatastore(nil, datastore.ErrNotFound)
	req := httptest.NewRequest(echo.GET, "/api/v1/photo/hash/size", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoGet(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
	// datastore error
	datastore.InitMokedDatastore(nil, errors.New("mocked"))
	req = httptest.NewRequest(echo.GET, "/api/v1/photo/hash/size", nil)
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoGet(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	// ok
	datastore.InitMokedDatastore([]byte{1, 2, 3}, nil)
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if assert.NoError(t, PhotoGet(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []byte{1, 2, 3}, rec.Body.Bytes())
	}
}

func TestPhotoPut(t *testing.T) {
	e := echo.New()

	// body not readable
	req := httptest.NewRequest(echo.PUT, "/api/v1/photo", errReader(0))
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "readBodyFailed", response.Code)
		}
	}

	// bad json
	req = httptest.NewRequest(echo.PUT, "/api/v1/photo", nil)
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "unmarshallBodyFailed", response.Code)
		}
	}

	// validation failed
	photoJson := []byte(`{"hash": "bar", "latitude":360}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/photo", bytes.NewBuffer(photoJson))
	c = middlewares.NewMockedContext(e.NewContext(req, rec))

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "errValidationFailed_6", response.Code)
		}
	}

	// not found
	photoJson = []byte(`{"hash": "bar"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/photo", bytes.NewBuffer(photoJson))
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "errNotFound", response.Code)
		}
	}

	// db err
	// not found
	photoJson = []byte(`{"hash": "bar"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/photo", bytes.NewBuffer(photoJson))
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "photoByHashFailed", response.Code)
		}
	}

	// update failed
	photoJson = []byte(`{"hash": "bar"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/photo", bytes.NewBuffer(photoJson))
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectQuery("^SELECT(.*)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked"))
	db.Mock.ExpectPrepare("^UPDATE photos (.*)").
		WillReturnError(errors.New("prepare error"))

	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "photoUpdateFailed", response.Code)
		}
	}
}

func TestPhotoDel(t *testing.T) {
	// not found
	e := echo.New()

	// Not found
	req := httptest.NewRequest(echo.DELETE, "/api/v1/photo", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnError(sql.ErrNoRows)
	if assert.NoError(t, PhotoDel(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "notFound", response.Code)
		}
	}

	// db error
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnError(errors.New("mocked"))
	if assert.NoError(t, PhotoDel(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "photoDeleteByHashFailed", response.Code)
		}
	}

	// ok
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	if err := datastore.InitMokedDatastore(nil, nil); err != nil {
		panic(err)
	}
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))

	if assert.NoError(t, PhotoDel(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := ApiResponseFromBody(rec.Body)
		if assert.NoError(t, err) {
			assert.True(t, response.Success)
		}
	}
}

func TestPhotoResize(t *testing.T) {
	e := echo.New()

	// height is not Atoiable
	req := httptest.NewRequest(echo.GET, "/api/v1/photo/hash/height/mocked", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))
	c.SetPath("/api/v1/photo/hash/height/:height")
	c.SetParamNames("height")
	c.SetParamValues("mocked")
	if assert.NoError(t, PhotoResize(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// height is not Atoiable
	c.SetPath("/api/v1/photo/id/width/:width")
	c.SetParamNames("width")
	c.SetParamValues("mocked")
	if assert.NoError(t, PhotoResize(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// height == width == 0
	c.SetPath("/api/v1/photo/id/width/:width")
	c.SetParamNames("width")
	c.SetParamValues("")
	if assert.NoError(t, PhotoResize(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// not found in datastore
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	c.SetParamNames("width")
	c.SetParamValues("100")

	if err := datastore.InitMokedDatastore(nil, errors.New("notfound")); err != nil {
		panic(err)
	}
	if assert.NoError(t, PhotoResize(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	// ok

}

func TestPhotoSearch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := middlewares.NewMockedContext(e.NewContext(req, rec))

	// error
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))
	if assert.NoError(t, PhotoSearch(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	// test "valid" hash
	rec = httptest.NewRecorder()
	c = middlewares.NewMockedContext(e.NewContext(req, rec))
	rows := sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(rows)
	if assert.NoError(t, PhotoSearch(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}
