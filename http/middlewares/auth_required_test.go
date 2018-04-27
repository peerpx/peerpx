package middlewares

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := AuthRequired()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// valid Key
	req.Header.Set("x-api-key", "QmU2TQthpXDj8QNK6jyqpWsjgDmr3E9Hn3F6zTahGGvZUC")
	assert.NoError(t, handler(ctx))

	// invalid key
	req.Header.Set("x-api-key", "dkfhsdjk")
	assert.Equal(t, http.StatusForbidden, handler(ctx).(*echo.HTTPError).Code)

}
