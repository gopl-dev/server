package app

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = ".config.yaml"

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

	Tracing struct {
		Enabled bool `yaml:"enabled"`
		// uptrace | log
		Driver string `yaml:"driver"`
		// https://uptrace.dev/
		UptraceDSN string `yaml:"uptrace_dsn"`
	} `yaml:"tracing"`

	Files struct {
		StorageDriver   string `yaml:"storage_driver"`
		MaxUploadSizeMB int64  `yaml:"max_upload_size_mb"`
		ImageMaxWidth   int    `yaml:"image_max_width"`
		ImageMaxHeight  int    `yaml:"image_max_height"`
		PreviewWidth    int    `yaml:"preview_width"`
		PreviewHeight   int    `yaml:"preview_height"`
		LocalFS         struct {
			StoragePath string `yaml:"storage_path"`
		} `yaml:"local_fs"`
	} `yaml:"files"`

	Entities struct {
		Books struct {
			Covers struct {
				Path string `yaml:"path"`
			} `yaml:"covers"`
		} `yaml:"books"`
	} `yaml:"entities"`

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

	OpenAPI struct {
		Enabled   bool   `yaml:"enabled"`
		ServePath string `yaml:"serve_path"`
	} `yaml:"openapi"`

	GoogleOAuth struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"google_oauth"`

	GithubOAuth struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"github_oauth"`

	// Admins is a list of user IDs with administrative privileges.
	// This is a temporary solution until  ACL is implemented.
	Admins []string `yaml:"admins"`
}

// IsDevEnv ...
func (c *ConfigT) IsDevEnv() bool {
	return c.App.Env == DevEnv
}

// IsTestEnv ...
func (c *ConfigT) IsTestEnv() bool {
	return c.App.Env == TestEnv
}

// IsProductionEnv ...
func (c *ConfigT) IsProductionEnv() bool {
	return c.App.Env == ProductionEnv
}

// TracingDisabled ...
func (c *ConfigT) TracingDisabled() bool {
	return !c.Tracing.Enabled
}

var loadConfigOnce sync.Once
var conf *ConfigT

// Config returns config from .config.yaml.
func Config() *ConfigT {
	loadConfigOnce.Do(func() {
		var err error
		conf, err = ConfigFromFile(defaultConfigFile)
		if err != nil {
			panic(err)
		}

		serverURL, err = url.Parse(conf.Server.Addr)
		if err != nil {
			panic(err)
		}
	})

	return conf
}

// ConfigFromFile returns new config from given YAML file.
func ConfigFromFile(filename string) (*ConfigT, error) {
	data, err := os.ReadFile(filename) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read '%s': %w", filename, err)
	}

	fileConf := new(ConfigT)
	err = yaml.Unmarshal(data, fileConf)
	if err != nil {
		return nil, err
	}

	return fileConf, err
}
