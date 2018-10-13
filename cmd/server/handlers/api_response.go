package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/peerpx/peerpx/cmd/server/context"
)

// APIResponse is the response returned by PeerPx API
type APIResponse struct {
	context    *context.AppContext
	clientIP   string
	UUID       string          `json:"uuid"`
	Timestamp  time.Time       `json:"timestamp"`
	HTTPStatus int             `json:"http_status"`
	Code       string          `json:"code"`
	Success    bool            `json:"success"`
	Message    string          `json:"message"` // message to user
	Log        string          `json:"-"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// NewAPIResponse return an instantiated API response
func NewAPIResponse(c *context.AppContext) *APIResponse {
	return &APIResponse{
		context:    c,
		UUID:       c.UUID,
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusOK,
		Data:       nil,
	}
}

// APIResponseFromBody unmarshall api response from http response body
// mainly used for tests
func APIResponseFromBody(body *bytes.Buffer) (APIResponse, error) {
	var response APIResponse
	err := json.Unmarshal(body.Bytes(), &response)
	return response, err
}

// Send send a reply
func (r *APIResponse) Send() error {
	r.Timestamp = time.Now()
	return r.context.JSON(r.HTTPStatus, r)
}

// OK send an OK response
func (r *APIResponse) OK(HTTPStatus int) error {
	r.Success = true
	r.HTTPStatus = HTTPStatus
	if r.Log != "" {
		r.context.LogInfo(r.Log)
	}
	return r.Send()
}

// KO send a not OK response (no error but success == false)
func (r *APIResponse) KO(HTTPStatus int) error {
	r.Success = false
	r.HTTPStatus = HTTPStatus
	if r.Log != "" {
		r.context.LogError(r.Log)
	}
	return r.Send()
}
