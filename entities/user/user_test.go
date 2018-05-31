package user

import (
	"testing"

	"github.com/peerpx/peerpx/services/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestUserCreate(t *testing.T) {
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
