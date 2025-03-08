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
	}

	Server struct {
		Port        string
		Host        string
		ApiBasePath string `yaml:"api_base_path"`
	}

	DB struct {
		Host       string
		Port       string
		User       string
		Password   string
		Name       string
		LogQueries bool `yaml:"log_queries"`
	}

	Content struct {
		Repo struct {
			// Path represents "{account}/{repo}" on GitHub.
			// Note that this is equivalent to "full_name" in the GitHub API.
			Path   string `yaml:"name"`
			Branch string `yaml:"branch"`
			Secret string `yaml:"secret"`
		}
		LocalDir string `yaml:"local_dir"`
	}
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
		err = yaml.Unmarshal(data, &conf)
		if err != nil {
			panic(err)
		}
	})

	return conf
}
