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
