package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host        string `yaml:"host"`
	PublicPort  string `yaml:"publicport"`
	PrivatePort string `yaml:"privateport"`
	DB          string `yaml:"database"`
	Login       string `yaml:"login"`
	Password    string `yaml:"password"`
}

type Database struct {
	DSN string `yaml:"dsn"` // "postgres://user:password@localhost:5432/dbname"
}

func NewConfig(configPath string) (*Config, error) {
	var config = new(Config)
	file, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Errorf("file error %w", err)
		return nil, err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
