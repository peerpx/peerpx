package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/peerpx/peerpx/cmd/server/context"
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

// ApiResponseFromBody unmarshall api response from http response body
// mainly used for tests
func ApiResponseFromBody(body *bytes.Buffer) (ApiResponse, error) {
	var response ApiResponse
	err := json.Unmarshal(body.Bytes(), &response)
	return response, err
}

// Send send a reply
func (r *ApiResponse) Send(c *context.AppContext, httpStatus int, code, message string) error {
	r.Timestamp = time.Now()
	r.HttpStatus = httpStatus
	r.Message = message
	r.Code = code
	return c.JSON(httpStatus, r)
}

// Error send an error reply
func (r *ApiResponse) Error(c *context.AppContext, httpStatus int, code, message string) error {
	r.Success = false
	if message != "" {
		log.Error(message)
	}
	return r.Send(c, httpStatus, code, message)
}

// OK send an OK response
func (r *ApiResponse) OK(c *context.AppContext, httpStatus int) error {
	r.Success = true
	return r.Send(c, httpStatus, "", "")
}
