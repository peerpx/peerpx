package middlewares

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAuthRequired(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := &AppContext{e.NewContext(req, rec), sessions.NewCookieStore([]byte("cookieAuthKey"), []byte("cookieEncryptionKey"))}

	handler := AuthRequired()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Via key in header
	// valid Key
	req.Header.Set("x-api-key", "QmU2TQthpXDj8QNK6jyqpWsjgDmr3E9Hn3F6zTahGGvZUC")
	if assert.NoError(t, handler(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// invalid key
	req.Header.Set("x-api-key", "dkfhsdjk")
	assert.Equal(t, http.StatusForbidden, handler(ctx).(*echo.HTTPError).Code)

	// test cookie auth
	mock := db.InitMockedDB("sqlmock_TestAuthRequired_1")
	defer db.DB.Close()
	rows := sqlmock.NewRows([]string{"id", "username", "password"}).AddRow(1, "toorop", "bla")
	mock.ExpectQuery("^SELECT(.*)").WillReturnRows(rows)
	ctx.SessionSet("username", "toorop")
	req.Header.Del("x-api-key")
	assert.NoError(t, handler(ctx))
	// no such user
	mock.ExpectQuery("^SELECT(.*)").WillReturnError(gorm.ErrRecordNotFound)
	assert.Equal(t, "record not found", handler(ctx).Error())

}
