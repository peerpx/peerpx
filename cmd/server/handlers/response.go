package handlers

import (
	"encoding/json"
	"time"

	"net/http"

	"bytes"

	"github.com/satori/go.uuid"
)

// ApiResponse is the response returned by peeepx API
type ApiResponse struct {
	UUID       string          `json:"uuid,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	HttpStatus int             `json:"http_status"`
	Code       string          `json:"code,omitempty"`
	Success    bool            `json:"success"`
	Message    string          `json:"message,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// NewApiResponse return an instantiated API response
// panic if uuid.NewV4() failed (should never happen)
func NewApiResponse() *ApiResponse {
	return &ApiResponse{
		Timestamp:  time.Now(),
		UUID:       uuid.Must(uuid.NewV4()).String(),
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
