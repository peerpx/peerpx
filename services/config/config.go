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
	isSet(key string) (bool, error)
	getE(key string) (interface{}, error)
	getIntE(key string) (int, error)
	getFloat64E(key string) (float64, error)
	getBoolE(key string) (bool, error)
	getStringE(key string) (string, error)
	getStringSliceE(key string) ([]string, error)
	getTimeE(key string) (time.Time, error)
	getDurationE(key string) (time.Duration, error)
}

// Interface

// Set set or update a value referenced buy key
func Set(key string, value interface{}) error {
	if conf == nil {
		return ErrNotInitialized
	}
	return conf.set(key, value)
}

// IsSet returns whether or not a key is associated with a value
func IsSet(key string) (bool, error) {
	if conf == nil {
		panic(ErrNotInitialized)
	}
	return conf.isSet(key)
}

// Get the value associated with the key as an interface and an error
func GetE(key string) (interface{}, error) {
	if conf == nil {
		return nil, ErrNotInitialized
	}
	return conf.getE(key)
}

// Get the value associated with the key as an interface
func Get(key string) interface{} {
	v, err := GetE(key)
	if err != nil {
		return nil
	}
	return v
}

// Get the value associated with the key as an interface if exists
// return  defaultValue on error or if not found
func GetDefault(key string, defaultValue interface{}) interface{} {
	v, err := GetE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetP returns the value associated with the key
// as an interface or panic if key is not found
func GetP(key string) interface{} {
	v, err := GetE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// int

// GetInt the value associated with the key as an integer and an error
func GetIntE(key string) (int, error) {
	if conf == nil {
		return 0, ErrNotInitialized
	}
	return conf.getIntE(key)
}

// GetInt the value associated with the key as an integer
func GetInt(key string) int {
	v, err := GetIntE(key)
	if err != nil {
		return 0
	}
	return v
}

// GetInt the value associated with the key as an integer
func GetIntDefault(key string, defaultValue int) int {
	v, err := GetIntE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetIntP returns the value associated with the key as an integer
// or panic if the key if not found or if value can't be converted
// to int
func GetIntP(key string) int {
	v, err := GetIntE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// float64

// GetFloat64E returns the value associated with the key as an float64 and an error
func GetFloat64E(key string) (float64, error) {
	if conf == nil {
		return 0, ErrNotInitialized
	}
	return conf.getFloat64E(key)
}

// GetFloat64 returns the value associated with the key as an float64
func GetFloat64(key string) float64 {
	v, err := GetFloat64E(key)
	if err != nil {
		return 0
	}
	return v
}

// GetFloat64Default return the value associated with the key as an float64
// or defaultValue if an error occurred, or if key is not found
func GetFloat64Default(key string, defaultValue float64) float64 {
	v, err := GetFloat64E(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetFloat64P returns the value associated with the key as an float64
// or panic if the key if not found or if value can't be converted
// to int
func GetFloat64P(key string) float64 {
	v, err := GetFloat64E(key)
	if err != nil {
		panic(err)
	}
	return v
}

// bool

// GetBoolE returns the value associated with the key as a boolean and an error
func GetBoolE(key string) (bool, error) {
	if conf == nil {
		return false, ErrNotInitialized
	}
	return conf.getBoolE(key)
}

// GetBool returns the value associated with the key as an bool
func GetBool(key string) bool {
	v, err := GetBoolE(key)
	if err != nil {
		return false
	}
	return v
}

// GetBoolDefault return the value associated with the key as an bool
// or defaultValue if an error occured, or if key is not found
func GetBoolDefault(key string, defaultValue bool) bool {
	v, err := GetBoolE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetBoolP returns the value associated with the key as an boolean
// or panic if the key if not found or if value can't be converted
// to int
func GetBoolP(key string) bool {
	v, err := GetBoolE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// string

// GetStringE returns the value associated with the key as a string and an error
func GetStringE(key string) (string, error) {
	if conf == nil {
		return "", ErrNotInitialized
	}
	return conf.getStringE(key)
}

// GetString returns the value associated with the key as an string
func GetString(key string) string {
	v, err := GetStringE(key)
	if err != nil {
		return ""
	}
	return v
}

// GetStringDefault return the value associated with the key as an string
// or defaultValue if an error occured, or if key is not found
func GetStringDefault(key string, defaultValue string) string {
	v, err := GetStringE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetStringP returns the value associated with the key as a string
// or panic if the key if not found or if value can't be converted
// to int
func GetStringP(key string) string {
	v, err := GetStringE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// string slice

// GetStringSliceE returns the value associated with the key as a string
// slice and an error
func GetStringSliceE(key string) ([]string, error) {
	if conf == nil {
		return nil, ErrNotInitialized
	}
	return conf.getStringSliceE(key)
}

// GetStringSlice returns the value associated with the key as an string slice
func GetStringSlice(key string) []string {
	v, err := GetStringSliceE(key)
	if err != nil {
		return nil
	}
	return v
}

// GetStringSliceDefault return the value associated with the key as an string
// slice or defaultValue if an error occurred, or if key is not found
func GetStringSliceDefault(key string, defaultValue []string) []string {
	v, err := GetStringSliceE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetStringSliceP returns the value associated with the key as a string
// slice or panic if the key if not found or if value can't be converted
// to a string slice
func GetStringSliceP(key string) []string {
	v, err := GetStringSliceE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// time

// GetTimeE returns the value associated with the key as a time.Time and an error
func GetTimeE(key string) (time.Time, error) {
	if conf == nil {
		return time.Time{}, ErrNotInitialized
	}
	return conf.getTimeE(key)
}

// GetTime returns the value associated with the key as an time.Time or zero value
func GetTime(key string) time.Time {
	v, err := GetTimeE(key)
	if err != nil {
		return time.Time{}
	}
	return v
}

// GetTimeDefault return the value associated with the key as a time.Time
// or defaultValue if an error occurred, or if key is not found
func GetTimeDefault(key string, defaultValue time.Time) time.Time {
	v, err := GetTimeE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetTimeP returns the value associated with the key as a time.Time
// or panic if the key if not found or if value can't be converted
// to a time.Time
func GetTimeP(key string) time.Time {
	v, err := GetTimeE(key)
	if err != nil {
		panic(err)
	}
	return v
}

// duration
// GetDurationE returns the value associated with the key as a time.Duration and an error
func GetDurationE(key string) (time.Duration, error) {
	if conf == nil {
		return 0, ErrNotInitialized
	}
	return conf.getDurationE(key)
}

// GetDuration returns the value associated with the key as an time.Duration or zero value
func GetDuration(key string) time.Duration {
	v, err := GetDurationE(key)
	if err != nil {
		return 0
	}
	return v
}

// GetDurationDefault return the value associated with the key as a time.Duration
// or defaultValue if an error occurred, or if key is not found
func GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	v, err := GetDurationE(key)
	if err != nil {
		return defaultValue
	}
	return v
}

// GetDurationP returns the value associated with the key as a time.Duration
// or panic if the key if not found or if value can't be converted
// to a time.Duration
func GetDurationP(key string) time.Duration {
	v, err := GetDurationE(key)
	if err != nil {
		panic(err)
	}
	return v
}
