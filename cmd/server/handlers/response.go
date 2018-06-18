package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/services/log"
)

// ApiResponse is the response returned by PeerPx API
type ApiResponse struct {
	UUID       string          `json:"uuid"`
	Timestamp  time.Time       `json:"timestamp"`
	HttpStatus int             `json:"http_status"`
	Code       string          `json:"code"`
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// NewApiResponse return an instantiated API response
// panic if uuid.NewV4() failed (should never happen)
func NewApiResponse(uuid string) *ApiResponse {
	return &ApiResponse{
		Timestamp:  time.Now(),
		UUID:       uuid,
		HttpStatus: http.StatusOK,
		Data:       nil,
	}
}

// GetApiResponse unmarshall api response from http response body
// mainly used for tests
func ApiResponseFromBody(body *bytes.Buffer) (ApiResponse, error) {
	var response ApiResponse
	err := json.Unmarshal(body.Bytes(), &response)
	return response, err
}

func (r *ApiResponse) Send(c *middlewares.AppContext, httpStatus int, code, message string, data json.RawMessage) error {
	r.Timestamp = time.Now()
	r.HttpStatus = httpStatus
	r.Message = message
	r.Code = code
	return c.JSON(httpStatus, r)
}

func (r *ApiResponse) Error(c *middlewares.AppContext, httpStatus int, code, message string) error {
	r.Success = false
	if message != "" {
		log.Error(message)
	}
	return r.Send(c, httpStatus, code, message, nil)
}

func (r *ApiResponse) OK(c *middlewares.AppContext, httpStatus int, data json.RawMessage) error {
	r.Success = true
	return r.Send(c, httpStatus, "", "", data)
}
