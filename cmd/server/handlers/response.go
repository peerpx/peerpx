package handlers

import "encoding/json"

// ApiResponse represents API response for all request
type ApiResponse struct {
	HttpStatus  int             `json:"status"`
	Code        string          `json:"code,omitempty"`
	Success     bool            `json:"success"`
	DevMessage  string          `json:"devmessage,omitempty"`
	UserMessage string          `json:"user_message, omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}
