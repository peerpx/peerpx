package user

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/services/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCreate(t *testing.T) {
	// bad email
	_, err := Create("foo", "john", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "foo is not a valid email", err.Error())
	}

	// username length
	viper.Set("usernameMaxLength", 5)
	viper.Set("usernameMinLength", 3)
	_, err = Create("foo@bar.com", "jojoletaxi", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "username must have 5 char max", err.Error())
	}
	_, err = Create("foo@bar.com", "jo", "blablabla")
	if assert.Error(t, err) {
		assert.Equal(t, "username must have 3 char min", err.Error())
	}

	// password length
	viper.Set("passwordMinLength", 6)
	_, err = Create("foo@bar.com", "jojo", "bla")
	assert.EqualError(t, err, "password must be at least 6 char long")

	// good
	mock := db.InitMockedDB("sqlmock_db_usercreate")
	defer db.DB.Close()
	mock.ExpectExec("^INSERT INTO \"users\"(.*)").WillReturnResult(sqlmock.NewResult(1, 1))
	user, err := Create("FOo@Bar.com", "jojo", "blablabla")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "foo@bar.com", user.Email)
		assert.Equal(t, "jojo", user.Username)
	}
}

func TestLogin(t *testing.T) {
	// by mail
	mock := db.InitMockedDB("sqlmock_db_userlogin")
	defer db.DB.Close()
	row := sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err := Login("john@doe.com", "secret")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "john@doe.com", user.Email)
		assert.Equal(t, "john", user.Username)
	}

	// bu username
	row = sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err = Login("john", "secret")
	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "john@doe.com", user.Email)
		assert.Equal(t, "john", user.Username)
	}

	row = sqlmock.NewRows([]string{"id", "username", "email", "password"}).AddRow(1, "john", "john@doe.com", "$2y$10$vjxV/XuyPaPdfdfuINLopc49COmFfxEiVFac4m0L7GgqvJ.KAQcfpmvCa")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(row)
	user, err = Login("john", "secret")
	assert.EqualError(t, err, "no such user")

	// ErrNoSuchUser
	mock.ExpectQuery("^SELECT(.*)").WillReturnError(gorm.ErrRecordNotFound)
	user, err = Login("john", "secret")
	assert.EqualError(t, err, "no such user")

}
