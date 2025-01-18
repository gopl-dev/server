package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DevEnv     = "DEV"
	TestEnv    = "TEST"
	StagingEnv = "STAGING"
	ReleaseEnv = "RELEASE"
)

type Config struct {
	App struct {
		ID      string `yaml:"id"`
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Env     string `yaml:"env"`
	}

	Server struct {
		Port        string
		Host        string
		ApiBasePath string `yaml:"api_base_path"`
	}

	DB struct {
		Addr       string
		User       string
		Password   string
		Name       string
		LogQueries bool `yaml:"log_queries"`
	}
}

var loadOnce sync.Once
var conf *Config

// Get returns config from .config.yaml
func Get() *Config {
	loadOnce.Do(func() {
		name := ".config.yaml"
		data, err := os.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("read '.config.yaml': %v", err)
			panic(err)
		}
		err = yaml.Unmarshal(data, &conf)
		if err != nil {
			panic(err)
		}
	})

	return conf
}

func IsDevEnv() bool {
	return conf.App.Env == DevEnv
}

func IsTestEnv() bool {
	return strings.EqualFold(conf.App.Env, TestEnv)
}

func IsReleaseEnv() bool {
	return strings.EqualFold(conf.App.Env, ReleaseEnv)
}
