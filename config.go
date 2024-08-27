package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type PortMapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Config struct {
	Port     int           `yaml:"port"`
	Mappings []PortMapping `yaml:"mappings"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Port == 0 {
		config.Port = 8000 // Default HTTP port if not specified
	}

	return &config, nil
}
