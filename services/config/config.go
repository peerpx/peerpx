package config

import (
	"errors"
	"time"
)

var conf Config

// errors
var ErrNotInitialized = errors.New("config: service not initialized")

// Config is the config service interface
type Config interface {
	set(key string, value interface{}) error
	get(key string) interface{}
	getOrPanic(key string) interface{}
	getInt(key string) int
	getIntOrPanic(key string) int
	getFloat64(key string) float64
	getFloat64OrPanic(key string) float64

	getBool(key string) bool
	getBoolOrPanic(key string) bool
	getString(key string) string
	getStringOrPanic(key string) string
	getStringSlice(key string) []string
	getStringSliceOrPanic(key string) []string
	getTime(key string) time.Time
	getTimeOrPanic(key string) time.Time
	getDuration(key string) time.Duration
	getDurationOrPanic(key string) time.Duration
	isSet(key string) bool
}

// Set set or update a value referenced buy key
func Set(key string, value interface{}) error {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.set(key, value)
}

// Get the value associated with the key as an interface
func Get(key string) interface{} {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.get(key)
}

// GetOrPanic returns the value associated with the key
// as an interface or panic if key is not found
func GetOrPanic(key string) interface{} {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getOrPanic(key)
}

// GetInt the value associated with the key as an integer
func GetInt(key string) int {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getInt(key)
}

// GetIntOrPanic the value associated with the key as an integer
// or panic if key if not found or if value can't be converted
// to int
func GetIntOrPanic(key string) int {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getIntOrPanic(key)
}

// GetFloat64 returns the value associated with the key as a float64
func GetFloat64(key string) float64 {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getFloat64(key)
}

// GetFloat64 returns the value associated with the key as a float64
// or if key if not found or if value can't be converted
//// to float64
func GetFloat64OrPanic(key string) float64 {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getFloat64OrPanic(key)
}

// GetBool returns the value associated with the key as a boolean
func GetBool(key string) bool {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getBool(key)
}

// GetBoolOrPanic returns the value associated with the key as a boolean
// or panic if key is not founs or if value can't be converted to a boolean
func GetBoolOrPanic(key string) bool {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getBoolOrPanic(key)
}

// GetString returns the value associated with the key as a string
func GetString(key string) string {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getString(key)
}

// GetStringOrPanic returns the value associated with the key as a boolean
// or panic if key is not found
func GetStringOrPanic(key string) string {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getStringOrPanic(key)
}

// GetStringSlice returns the value associated with the key as a string slice
func GetStringSlice(key string) []string {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getStringSlice(key)
}

// GetStringSliceOrPanic returns the value associated with the key as a slice
// of string or panic if key is not found
func GetStringSliceOrPanic(key string) []string {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getStringSliceOrPanic(key)
}

// GetTime returns the value associated with the key as a time.Time
func GetTime(key string) time.Time {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getTime(key)
}

// GetTimeOrPanic returns the value associated with the key as a time.Time
// or panic if key is not found or if value can't be converted to time.Time
func GetTimeOrPanic(key string) time.Time {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getTimeOrPanic(key)
}

// GetDuration returns the value associated with the key as a time.Duration
func GetDuration(key string) (bool, time.Duration) {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getDuration(key)
}

// GetDuration returns the value associated with the key as a time.Duration
// panic if key is not fond or if value can(t be converted to time.Duration
func GetDurationOrPanic(key string) time.Duration {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.getDurationOrPanic(key)
}

// IsSet returns whether or not a key is associated with a value
func IsSet(key string) bool {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.isSet(key)
}
