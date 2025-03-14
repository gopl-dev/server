package app

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

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
		ApiBasePath string `yaml:"api_base_path"`
	} `yaml:"server"`

	DB struct {
		Host       string `yaml:"host"`
		Port       string `yaml:"port"`
		User       string `yaml:"user"`
		Password   string `yaml:"password"`
		Name       string `yaml:"name"`
		LogQueries bool   `yaml:"log_queries"`
	} `yaml:"db"`
}

func (c ConfigT) IsDevEnv() bool {
	return c.App.Env == DevEnv
}

func (c ConfigT) IsTestEnv() bool {
	return c.App.Env == TestEnv
}

func (c ConfigT) IsReleaseEnv() bool {
	return c.App.Env == ReleaseEnv
}

var loadConfigOnce sync.Once
var conf *ConfigT

// Config returns config from .config.yaml
func Config() *ConfigT {
	loadConfigOnce.Do(func() {
		name := ".config.yaml"
		data, err := os.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("read '.config.yaml': %v", err)
			panic(err)
		}

		conf = &ConfigT{}
		err = yaml.Unmarshal(data, conf)
		if err != nil {
			panic(err)
		}
	})

	return conf
}
