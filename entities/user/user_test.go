package user

import (
	"testing"

	"strings"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func init() {
	db.InitMockedDatabase()
}

func TestCreate(t *testing.T) {
	config.InitBasicConfig(strings.NewReader(""))
	// bad email
	_, err := Create("foo", "john", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "foo is not a valid email", err.Error())
	}

	// username length
	config.Set("username.maxLength", "5")
	config.Set("username.minLength", "3")

	_, err = Create("foo@bar.com", "jojoletaxi", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "username must have 5 char max", err.Error())
	}
	_, err = Create("foo@bar.com", "jo", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "username must have 3 char min", err.Error())
	}

	// password length
	config.Set("password.minLength", "6")
	_, err = Create("foo@bar.com", "jojo", "bla")
	assert.EqualError(t, err, "password must be at least 6 char long")

	db.Mock.ExpectPrepare("^INSERT INTO users (.*)").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))
	user, err := Create("FOo@Bar.com", "jojo", "blablabla")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "foo@bar.com", user.Email)
		assert.Equal(t, "jojo", user.Username)
	}
}

func TestLogin(t *testing.T) {
	// by mail
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err := Login("john@doe.com", "secret")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "john@doe.com", user.Email)
		assert.Equal(t, "john", user.Username)
	}

	// bu username
	row = sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err = Login("john", "secret")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "john@doe.com", user.Email)
		assert.Equal(t, "john", user.Username)
	}

	row = sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPdfdfuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err = Login("john", "secret")
	assert.EqualError(t, err, "no such user")

	// ErrNoSuchUser
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(gorm.ErrRecordNotFound)
	user, err = Login("john", "secret")
	assert.EqualError(t, err, "no such user")
}
