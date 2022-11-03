package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Host           string
	DBPort         int
	User           string
	Password       string
	DBname         string
	Port           string
	ConnectionType string
}

func GetConfig(path string) (*Config, error) {
	var err error
	var conf *Config

	_, err = toml.DecodeFile(path, &conf)

	return conf, err
}
