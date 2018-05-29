package config

import "time"

type Config interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool
}

func Get(key string) interface{} {
	return nil
}
func GetBool(key string) bool { return false }

func GetFloat64(key string) float64 {
	return 0
}

func GetInt(key string) int { return 0 }

func GetString(key string) string {
	return ""
}

func GetStringMap(key string) (strMap map[string]interface{}) {
	return

}

func GetStringMapString(key string) (strStrMapmap [string]string) {
	return
}

func GetStringSlice(key string) []string {
	return []string{}
}

func GetTime(key string) time.Time {
	return time.Now()
}

func GetDuration(key string) (duration time.Duration) {
	return
}
func IsSet(key string) bool {
	return false
}
