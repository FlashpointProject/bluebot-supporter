package main

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

// Config is config
type Config struct {
	AccessKeys []AccessKey
	Port       int
}

// AccessKey is access key
type AccessKey struct {
	UserID  string
	UserKey string
}

func loadConfig(l *logrus.Logger) *Config {
	file, err := os.Open("config.json")
	panicErr(l, err)
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	panicErr(l, err)

	return config
}
