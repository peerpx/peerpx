package photo

import (
	"testing"

	"time"

	"errors"

	"database/sql"

	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func init() {
	db.InitMockedDatabase()
}

func TestGetByHash(t *testing.T) {
	row := sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	photo, err := GetByHash("mocked")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), photo.ID)
		assert.Equal(t, "mocked", photo.Hash)
	}
}

func TestDeleteByHash(t *testing.T) {
	// prepare failed
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").WillReturnError(errors.New("mocked error"))
	err := DeleteByHash("foo")
	assert.EqualError(t, err, "mocked error")

	// not found
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnError(sql.ErrNoRows)
	err = DeleteByHash("foo")
	assert.EqualError(t, err, sql.ErrNoRows.Error())

	// error on datastore delete
	if err = datastore.InitMokedDatastore(nil, errors.New("mocked")); err != nil {
		panic(err)
	}
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	err = DeleteByHash("foo")
	assert.EqualError(t, err, "mocked")

	//not found in data store must returns nil error
	if err = datastore.InitMokedDatastore(nil, datastore.ErrNotFound); err != nil {
		panic(err)
	}
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	err = DeleteByHash("foo")
	assert.NoError(t, err)

	// OK
	if err = datastore.InitMokedDatastore(nil, nil); err != nil {
		panic(err)
	}
	db.Mock.ExpectPrepare("^DELETE FROM photos (.*)").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	err = DeleteByHash("foo")
	assert.NoError(t, err)

}

func TestList(t *testing.T) {
	row := sqlmock.NewRows([]string{"id", "hash"}).AddRow(1, "mocked").AddRow(2, "mocked2")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	photos, err := List("foo", "bar")
	if assert.NoError(t, err) {
		assert.Equal(t, 2, len(photos))
		assert.Equal(t, uint(1), photos[0].ID)
		assert.Equal(t, "mocked", photos[0].Hash)
		assert.Equal(t, uint(2), photos[1].ID)
		assert.Equal(t, "mocked2", photos[1].Hash)
	}
}

func TestPhoto_Create(t *testing.T) {
	photo := new(Photo)
	// prepare failed
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		WillReturnError(errors.New("prepare error"))
	err := photo.Create()
	assert.EqualError(t, err, "prepare error")

	// exec failed
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		ExpectExec().
		WillReturnError(errors.New("prepare error"))
	err = photo.Create()
	assert.EqualError(t, err, "prepare error")

	// OK
	db.Mock.ExpectPrepare("^INSERT INTO photos (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	err = photo.Create()
	if assert.NoError(t, err) {
		assert.Equal(t, photo.ID, uint(1))
	}
}

func TestPhoto_Update(t *testing.T) {
	photo := new(Photo)
	err := photo.Update()
	assert.EqualError(t, err, "photo is not recoded in DB yet, i can't update it !")

	photo.ID = 1
	// prepare failed
	db.Mock.ExpectPrepare("^UPDATE photos (.*)").
		WillReturnError(errors.New("prepare error"))
	err = photo.Update()
	assert.EqualError(t, err, "prepare error")

	// exec failed
	db.Mock.ExpectPrepare("^UPDATE photos (.*)").
		ExpectExec().
		WillReturnError(errors.New("prepare error"))
	err = photo.Update()
	assert.EqualError(t, err, "prepare error")

	// OK
	db.Mock.ExpectPrepare("^UPDATE photos (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	err = photo.Update()
	assert.NoError(t, err)
}

func TestPhoto_Validate(t *testing.T) {
	const longString = "6hRRCtSTRK53GQEItACt7Uryq90dVBZfoqOzNFOAb6F3SvS0kcUzRNpfBo7FONRubzDznAO9PlqN5yHr2HWK3gXNdZKAKw0e4fsEk4aSkc4eTPounHQwLmtQo8pyVGPsnpe8M5mwbRQSoj2rQlmmAhcCj1BtfbibF0UemN4Ya6DSibjyHyM8zKDXccVwmQ4ZbXHDC5XMsKIivoFga8EgHWCcQ0qrjSzBAilVwuUpNHoXumIOYqF1QOvGfCPLYW21"
	photo := new(Photo)

	// OK
	assert.Equal(t, uint8(0), photo.Validate())

	// Name
	photo.Name = longString
	assert.Equal(t, uint8(1), photo.Validate())
	photo.Name = ""

	// Camera
	photo.Camera = longString
	assert.Equal(t, uint8(2), photo.Validate())
	photo.Camera = ""

	// Lens
	photo.Lens = longString
	assert.Equal(t, uint8(3), photo.Validate())
	photo.Lens = ""

	// ShutterSpeed
	photo.ShutterSpeed = longString
	assert.Equal(t, uint8(4), photo.Validate())
	photo.ShutterSpeed = ""

	// Location
	photo.Location = longString
	assert.Equal(t, uint8(5), photo.Validate())
	photo.Location = ""

	// Latitude
	photo.Latitude = -90.01
	assert.Equal(t, uint8(6), photo.Validate())
	photo.Latitude = 90.01
	assert.Equal(t, uint8(6), photo.Validate())
	photo.Latitude = 0.00

	// Longitude
	photo.Longitude = -180.01
	assert.Equal(t, uint8(7), photo.Validate())
	photo.Longitude = 180.01
	assert.Equal(t, uint8(7), photo.Validate())
	photo.Longitude = 0.00

	// TakenAt
	photo.TakenAt = time.Now().Add(10 * time.Hour)
	assert.Equal(t, uint8(8), photo.Validate())
	photo.TakenAt = time.Now().Add(-10 * time.Hour)
}
