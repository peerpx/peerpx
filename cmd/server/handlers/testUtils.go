package handlers

import (
	"errors"
)

// test reader error ( body, err := ioutil.ReadAll(c.Request().Body)
type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked")
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
