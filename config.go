package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	Host string `mapstructure:"HOST"`
	Port string `mapstructure:"PORT"`

	DBUrl string `mapstructure:"DB_URL"`
}
 
func LoadConfig(v *viper.Viper, path string) (config Config) {
	v.AddConfigPath(".")
	v.SetConfigName(path)
	v.SetConfigType("env")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("config: %v\n", err)
		return
	}
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("config: %v\n", err)
		return
	}
	return
}
