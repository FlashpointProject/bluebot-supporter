package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/sirupsen/logrus"
)

// Config is config
type Config struct {
	Port int
}

// loadConfig reads CONFIG_PATH (default config.json). If CONFIG_PATH is not
// set, a missing config.json is fine and the defaults are used.
func loadConfig(l *logrus.Logger) *Config {
	config := &Config{Port: 80}

	path := os.Getenv("CONFIG_PATH")
	explicit := path != ""
	if !explicit {
		path = "config.json"
	}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) && !explicit {
		l.WithField("port", config.Port).Infoln("no config.json found, using defaults")
		return config
	}
	panicErr(l, err)
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	panicErr(l, err)

	return config
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
