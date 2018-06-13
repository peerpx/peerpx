package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
)

// GetApiResponse unmarshall api response panic
func GetApiResponse(body *bytes.Buffer) (ApiResponse, error) {
	var response ApiResponse
	err := json.Unmarshal(body.Bytes(), &response)
	return response, err

}

// test reader error ( body, err := ioutil.ReadAll(c.Request().Body)
type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked")
}
