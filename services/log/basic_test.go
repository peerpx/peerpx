package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	InitBasicLogger(buf)
	Info("foo", "bar")
	assert.True(t, strings.HasSuffix(buf.String(), "info: foo bar\n"))
	Infof("foo is %s", "bar")
	assert.True(t, strings.HasSuffix(buf.String(), "info: foo is bar\n"))
	Error("foo", "bar")
	assert.True(t, strings.HasSuffix(buf.String(), "error: foo bar\n"))
	Errorf("foo is %s", "bar")
	assert.True(t, strings.HasSuffix(buf.String(), "error: foo is bar\n"))
}
