package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	DBHost         string `yaml:"dbhost"`
	DBPort         int    `yaml:"dbport"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DBname         string `yaml:"dbname"`
	ConnectionType string `yaml:"connectiontype"`
	ContexTimeout  int    `yaml:"contextimeout"`
	MigrationPath  string `yaml:"migrationpath"`
	ReadTimeout    int    `yaml:"readtimeout"`
	WriteTimeout   int    `yaml:"writetimeout"`
}

func GetConfig(path string) (*Config, error) {

	conf := &Config{}
	in, err := os.Open(path)
	if err != nil {
		return conf, err
	}
	defer func() {
		if err := in.Close(); err != nil {
			log.Println(err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := in.Read(buf)
	if err != nil && err != io.EOF {
		log.Println(err)
	}

	err = yaml.Unmarshal(buf[:n], conf)

	return conf, err
}
