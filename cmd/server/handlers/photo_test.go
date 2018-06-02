package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestPhotoPost(t *testing.T) {
	// init config (small values -> photo will be re-encoded)
	config.InitBasicConfig(strings.NewReader(""))
	config.Set("photo.maxWidth", "100")
	config.Set("photo.maxHeight", "100")

	//  init mocked datastore
	datastore.InitMokedDatastore([]byte{}, nil)

	// mocked DB
	mock := db.InitMockedDB("sqlmock_db_0")
	defer db.DB.Close()
	mock.ExpectExec("^INSERT INTO \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))

	data := `{"Name":"ma super photo", "Description":" ma description"}`

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	handleErr(writer.WriteField("data", data))

	file, err := os.Open("../../../etc/samples/photos/robin.jpg")
	handleErr(err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)

	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config.InitBasicConfig(strings.NewReader(""))
	config.Set("hostname", "peerpx.com")
	if assert.NoError(t, PhotoPost(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		var response PhotoPostResponse
		err = json.Unmarshal(resp, &response)
		assert.NoError(t, err)
		assert.Equal(t, "AE2LNbBQ3vBDJHZpyNt3g9dFs14gcDryx6Lcted6d1yE", response.PhotoProps.Hash)
	}
}

func TestPhotoPostNotAPhoto(t *testing.T) {
	data := `{"Name":"ma super photo", "Description":" ma description"}`

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	handleErr(writer.WriteField("data", data))

	file, err := os.Open("../../../etc/samples/photos/not-a-photo.jpg")
	handleErr(err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", "robin.jpg")
	handleErr(err)

	_, err = io.Copy(part, file)
	handleErr(err)
	handleErr(writer.Close())

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
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

func TestPhotoGetPropertiesByHash(t *testing.T) {
	// mocked DB
	mock := db.InitMockedDB("sqlmock_db_1")
	defer db.DB.Close()

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
	mock := db.InitMockedDB("sqlmock_db_2")
	defer db.DB.Close()

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
	datastore.InitMokedDatastore([]byte{1, 2, 3}, nil)

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
	datastore.InitMokedDatastore(nil, datastore.ErrNotFound)

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

// TODO ?
func TestPhotoResize(t *testing.T) {
}

func TestPhotoDel(t *testing.T) {
	mock := db.InitMockedDB("sqlmock_db_3")
	mock.ExpectExec("DELETE FROM \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))
	defer db.DB.Close()
	datastore.InitMokedDatastore(nil, nil)
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
	mock := db.InitMockedDB("sqlmock_db_4")
	defer db.DB.Close()

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

func TestPhotoPut(t *testing.T) {
	// mocked DB
	mock := db.InitMockedDB("sqlmock_db_5")
	defer db.DB.Close()
	e := echo.New()
	rec := httptest.NewRecorder()

	// body doesn't represent a valid json strut
	req := httptest.NewRequest("PUT", "/", bytes.NewBuffer([]byte{}))
	c := e.NewContext(req, rec)
	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// no found (bad hash)
	photoJson := []byte(`{"hash": "bar"}`)
	mock.ExpectQuery("^SELECT(.*)").WillReturnError(gorm.ErrRecordNotFound)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/", bytes.NewBuffer(photoJson))
	c = e.NewContext(req, rec)
	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	// bad props should return error and empty photo
	photoJson = []byte(`{"hash": "bar", "latitude":360}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/", bytes.NewBuffer(photoJson))
	c = e.NewContext(req, rec)
	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var response PhotoPutResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if assert.NoError(t, err) {
			assert.Equal(t, uint8(6), response.Code)
		}
	}

	// returned
	photoNew := []byte(`{
	"hash": "should not be modified",
	"name": "newName",
	"description": "newDescription",
	"camera": "newCamera",
	"lens": "newLens",
	"focalLength": 50,
	"iso": 100,
	"shutterSpeed": "newShutterSpeed",
	"aperture": 0.6,
	"location": "newLocation",
	"privacy": false,
	"latitude": 80,
	"longitude": 80,
	"takenAt": "2018-05-21T15:04:05Z",
	"nsfw": false}
	`)

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "mocked")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(rows)
	mock.ExpectExec("^UPDATE \"photos\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/", bytes.NewBuffer(photoNew))
	c = e.NewContext(req, rec)
	if assert.NoError(t, PhotoPut(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// get body
		// json -> stuct
		var response PhotoPutResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if assert.NoError(t, err) {
			assert.Equal(t, uint8(0), response.Code)
			assert.Equal(t, "", response.Photo.Hash)
			assert.Equal(t, "newName", response.Photo.Name)
			assert.Equal(t, "newDescription", response.Photo.Description)
			assert.Equal(t, "newCamera", response.Photo.Camera)
			assert.Equal(t, "newLens", response.Photo.Lens)
			assert.Equal(t, uint16(50), response.Photo.FocalLength)
			assert.Equal(t, uint16(100), response.Photo.Iso)
			assert.Equal(t, "newShutterSpeed", response.Photo.ShutterSpeed)
			assert.Equal(t, float32(0.6), response.Photo.Aperture)
			assert.Equal(t, "newLocation", response.Photo.Location)
			assert.Equal(t, false, response.Photo.Privacy)
			assert.Equal(t, float32(80.00), response.Photo.Latitude)
			assert.Equal(t, float32(80.00), response.Photo.Longitude)
			assert.Equal(t, int64(1526915045), response.Photo.TakenAt.Unix())
			assert.Equal(t, false, response.Photo.Nsfw)
		}

	}

}
