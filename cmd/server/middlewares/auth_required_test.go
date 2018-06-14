package middlewares

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"database/sql"

	"errors"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func init() {
	db.InitMockedDatabase()
}

func TestAuthRequired(t *testing.T) {
	// no cookie store
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := &AppContext{e.NewContext(req, rec), nil, "mocked"}
	handler := AuthRequired()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})
	assert.Panics(t, func() { handler(ctx) })

	// no username in session -> forbidden
	ctx = NewMockedContext(e.NewContext(req, rec))
	err := handler(ctx)
	assert.Error(t, err, echo.ErrForbidden)

	// username in session but user not found in DB
	ctx.SessionSet("username", "toorop")
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(sql.ErrNoRows)
	err = handler(ctx)
	assert.EqualError(t, err, echo.ErrForbidden.Error())

	// err with DB
	db.Mock.ExpectQuery("^SELECT(.*)").WillReturnError(errors.New("mocked"))
	err = handler(ctx)
	assert.EqualError(t, err, "mocked")

	// ok
	db.Mock.ExpectQuery("^SELECT(.*)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "toorop"))
	err = handler(ctx)
	if assert.NoError(t, err) {
		user := ctx.Get("u").(*user.User)
		assert.Equal(t, uint(1), user.ID)
	}
}
