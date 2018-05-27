package config

import "github.com/spf13/viper"

// InitViper initialize viper
func InitViper() error {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	return viper.ReadInConfig()
}
