// Package config holds the configuraiton for the application
package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type YAMLURL struct {
	*url.URL
}

func (j *YAMLURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	url, err := url.Parse(s)
	j.URL = url
	return err
}

// Config is the configuration struct for the app.
type Config struct {
	LogLevel string `yaml:"log_level"`
	Gmetad   `yaml:"gmetad"`
}

// Gmetad is the configuration for Ganglia's gmetad process.
type Gmetad struct {
	URL   YAMLURL       `yaml:"url"`
	Sleep time.Duration `yaml:"sleep"`
}

// DefaultPaths contains the default paths used to search for a config file.
var DefaultPaths = []string{
	"/data/private/share/daemons/ct-metric-processing-daemon/config/config.yml",
	"./config/config.yml",
}

// FromFile parses the given file path and returns a Config.
func FromFile(paths []string) (*Config, error) {
	path, err := findConfigFile(paths)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func findConfigFile(paths []string) (string, error) {
	for _, path := range paths {
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("Unable to find config file")
}
