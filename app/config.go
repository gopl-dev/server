package app

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// ConfigT ...
type ConfigT struct {
	App struct {
		ID      string `yaml:"id"`
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Env     string `yaml:"env"`
	} `yaml:"app"`

	Server struct {
		Port        string `yaml:"port"`
		Host        string `yaml:"host"`
		APIBasePath string `yaml:"api_base_path"`
		Addr        string `yaml:"addr"`
	} `yaml:"server"`

	DB struct {
		Host       string `yaml:"host"`
		Port       string `yaml:"port"`
		User       string `yaml:"user"`
		Password   string `yaml:"password"`
		Name       string `yaml:"name"`
		LogQueries bool   `yaml:"log_queries"`
	} `yaml:"db"`

	Email struct {
		// Driver can be: smtp or test
		Driver   string `yaml:"driver"`
		From     string `yaml:"from"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"email"`

	Session struct {
		DurationHours int    `yaml:"duration_hours"`
		Key           string `yaml:"key"`
	} `yaml:"session"`
}

// IsDevEnv ...
func (c ConfigT) IsDevEnv() bool {
	return c.App.Env == DevEnv
}

// IsTestEnv ...
func (c ConfigT) IsTestEnv() bool {
	return c.App.Env == TestEnv
}

// IsReleaseEnv ...
func (c ConfigT) IsReleaseEnv() bool {
	return c.App.Env == ReleaseEnv
}

var loadConfigOnce sync.Once
var conf *ConfigT

// Config returns config from .config.yaml.
func Config() *ConfigT {
	loadConfigOnce.Do(func() {
		name := ".config.yaml"

		data, err := os.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("read '.config.yaml': %w", err)
			panic(err)
		}

		conf = new(ConfigT)

		err = yaml.Unmarshal(data, conf)
		if err != nil {
			panic(err)
		}
	})

	return conf
}
