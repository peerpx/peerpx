package models

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

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
