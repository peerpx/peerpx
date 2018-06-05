package config

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

func configTest(t *testing.T) {
	// Get* on valid key
	assert.Equal(t, "foo", Get("string").(string))
	v, err := GetE("string")
	if assert.NoError(t, err) {
		assert.Equal(t, "foo", v.(string))
	}
	assert.NotPanics(t, func() {
		GetP("string")
	})
	// Get on a key that does not exist
	assert.Equal(t, "", Get("noexist").(string))
	v, err = GetE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetP("noexist")
	})
	assert.Equal(t, "bar", GetDefault("noexists", "bar"))
	assert.Equal(t, "foo", GetDefault("string", "bar").(string))

	// int
	// Get* on valid key
	assert.Equal(t, 12, GetInt("int"))
	vint, err := GetIntE("int")
	if assert.NoError(t, err) {
		assert.Equal(t, 12, vint)
	}
	assert.NotPanics(t, func() {
		GetIntP("int")
	})
	// Get on a key that does not exist
	assert.Equal(t, 0, GetInt("noexist"))
	v, err = GetIntE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetIntP("noexist")
	})
	assert.Equal(t, 666, GetIntDefault("noexists", 666))
	assert.Equal(t, 12, GetIntDefault("int", 5))

	// not a parsable int
	_, err = GetIntE("invalidint")
	assert.Error(t, err)

	// float64
	assert.Equal(t, 1.2, GetFloat64("float64"))
	vfloat64, err := GetFloat64E("float64")
	if assert.NoError(t, err) {
		assert.Equal(t, 1.2, vfloat64)
	}
	assert.NotPanics(t, func() {
		GetFloat64P("float64")
	})
	// Get on a key that does not exist
	assert.Equal(t, 0.0, GetFloat64("noexist"))
	v, err = GetFloat64E("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetFloat64P("noexist")
	})
	assert.Equal(t, 6.66, GetFloat64Default("noexists", 6.66))
	assert.Equal(t, 1.2, GetFloat64Default("float64", 3.33))
	// not a parsable float64
	_, err = GetFloat64E("invalidfloat64")
	assert.Error(t, err)

	// bool
	assert.Equal(t, true, GetBool("bool"))
	vbool, err := GetBoolE("bool")
	if assert.NoError(t, err) {
		assert.Equal(t, true, vbool)
	}
	assert.NotPanics(t, func() {
		GetBoolP("bool")
	})
	// Get on a key that does not exist
	assert.Equal(t, false, GetBool("noexist"))
	v, err = GetBoolE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetBoolP("noexist")
	})
	assert.Equal(t, true, GetBoolDefault("noexists", true))
	assert.Equal(t, true, GetBoolDefault("bool", false))
	// invalid bool
	_, err = GetBoolE("invalidbool")
	assert.Error(t, err)

	// string
	assert.Equal(t, "foo", GetString("string"))
	vstring, err := GetStringE("string")
	if assert.NoError(t, err) {
		assert.Equal(t, "foo", vstring)
	}
	assert.NotPanics(t, func() {
		GetStringP("string")
	})
	// Get on a key that does not exist
	assert.Equal(t, "", GetString("noexist"))
	v, err = GetStringE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetStringP("noexist")
	})
	assert.Equal(t, "bar", GetStringDefault("noexists", "bar"))
	assert.Equal(t, "foo", GetStringDefault("string", "bar"))

	// string slice
	assert.Equal(t, []string{"foo", "bar", "back"}, GetStringSlice("stringslice"))
	vstringSlice, err := GetStringSliceE("stringslice")
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"foo", "bar", "back"}, vstringSlice)
	}
	assert.NotPanics(t, func() {
		GetStringSliceP("stringslice")
	})
	// Get on a key that does not exist
	assert.Equal(t, []string{}, GetStringSlice("noexist"))
	v, err = GetStringSliceE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetStringSliceP("noexist")
	})
	assert.Equal(t, []string{"bar", "back"}, GetStringSliceDefault("noexists", []string{"bar", "back"}))
	assert.Equal(t, []string{"foo", "bar", "back"}, GetStringSliceDefault("stringslice", []string{"bar", "back"}))
	// empty slice
	es, err := GetStringSliceE("empty")
	assert.NoError(t, err)
	assert.Equal(t, []string{}, es)

	// time
	goodTime := time.Unix(1528031462, 0)
	assert.Equal(t, goodTime, GetTime("time"))
	vtime, err := GetTimeE("time")
	if assert.NoError(t, err) {
		assert.Equal(t, goodTime, vtime)
	}
	assert.NotPanics(t, func() {
		GetTimeP("time")
	})
	// Get on a key that does not exist
	assert.Equal(t, time.Time{}, GetTime("noexist"))
	v, err = GetTimeE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetTimeP("noexist")
	})
	assert.Equal(t, goodTime, GetTimeDefault("noexists", goodTime))
	assert.Equal(t, goodTime, GetTimeDefault("time", time.Now()))
	// invalid time
	_, err = GetTimeE("invalidtime")
	assert.Error(t, err)

	// duration
	goodDuration, _ := time.ParseDuration("2h45m")
	assert.Equal(t, goodDuration, GetDuration("duration"))
	vDuration, err := GetDurationE("duration")
	if assert.NoError(t, err) {
		assert.Equal(t, goodDuration, vDuration)
	}
	assert.NotPanics(t, func() {
		GetDurationP("duration")
	})
	// Get on a key that does not exist
	assert.Equal(t, time.Duration(0), GetDuration("noexist"))
	v, err = GetDurationE("noexist")
	if assert.Error(t, err) {
		assert.EqualError(t, err, ErrNotFound.Error())
	}
	assert.Panics(t, func() {
		GetDurationP("noexist")
	})
	assert.Equal(t, goodDuration, GetDurationDefault("noexists", goodDuration))
	tduration, _ := time.ParseDuration("3m55s")
	assert.Equal(t, goodDuration, GetDurationDefault("duration", tduration))
	// invalid time
	_, err = GetDurationE("invaliduration")
	assert.Error(t, err)
}

func TestUnitialized(t *testing.T) {
	conf = nil
	assert.Error(t, Set("foo", "bar"))
	_, err := GetE("foo")
	assert.Error(t, err)
	_, err = GetIntE("foo")
	assert.Error(t, err)
	_, err = GetFloat64E("foo")
	assert.Error(t, err)
	_, err = GetBoolE("foo")
	assert.Error(t, err)
	_, err = GetStringE("foo")
	assert.Error(t, err)
	_, err = GetStringSliceE("foo")
	assert.Error(t, err)
	_, err = GetTimeE("foo")
	assert.Error(t, err)
	_, err = GetDurationE("foo")
	assert.Error(t, err)
	_, err = IsSet("foo")
	assert.Error(t, err)
}
