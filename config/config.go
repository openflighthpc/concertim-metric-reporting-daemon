//==============================================================================
// Copyright (C) 2024-present Alces Flight Ltd.
//
// This file is part of Concertim Metric Reporting Daemon.
//
// This program and the accompanying materials are made available under
// the terms of the Eclipse Public License 2.0 which is available at
// <https://www.eclipse.org/legal/epl-2.0>, or alternative license
// terms made available by Alces Flight Ltd - please direct inquiries
// about licensing to licensing@alces-flight.com.
//
// Concertim Metric Reporting Daemon is distributed in the hope that it will be useful, but
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
// IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
// OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
// PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
// details.
//
// You should have received a copy of the Eclipse Public License 2.0
// along with Concertim Metric Reporting Daemon. If not, see:
//
//  https://opensource.org/licenses/EPL-2.0
//
// For more information on Concertim Metric Reporting Daemon, please visit:
// https://github.com/openflighthpc/concertim-metric-reporting-daemon
//==============================================================================

// Package config holds the configuraiton for the application
package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is the configuration struct for the app.
type Config struct {
	LogLevel         string `yaml:"log_level"`
	LogFile          string `yaml:"log_file"`
	SharedSecretFile string `yaml:"shared_secret_file"`
	API              `yaml:"api"`
	DSM              `yaml:"dsm"`
	VisualizerAPI    `yaml:"visualizer_api"`
	RRD              `yaml:"rrd"`
}

// API is the configuration for the HTTP API component.
type API struct {
	IP           string        `yaml:"ip"`
	JWTSecret    []byte        `yaml:"-"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// DSM is the configuration for the Data Source Map component.
type DSM struct {
	Frequency time.Duration `yaml:"frequency"`
	Testdata  string        `yaml:"testdata"`
	Throttle  time.Duration `yaml:"throttle"`
}

type VisualizerAPI struct {
	AuthUrl              string `yaml:"auth_url"`
	DataSourceMapUrl     string `yaml:"data_source_map_url"`
	JWTSecret            []byte `yaml:"-"`
	Password             string `yaml:"password"`
	SkipCertificateCheck bool   `yaml:"skip_certificate_check"`
	Username             string `yaml:"username"`
}

type RRD struct {
	ClusterName string        `yaml:"cluster_name"`
	Directory   string        `yaml:"directory"`
	GridName    string        `yaml:"grid_name"`
	Step        time.Duration `yaml:"step"`
	ToolPath    string        `yaml:"rrd_tool_path"`
}

// DefaultPath is the path to the default config file.
const DefaultPath string = "/opt/concertim/opt/ct-metric-reporting-daemon/config/config.yml"

// FromFile parses the given file path and returns a Config.
func FromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "loading config")
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("loading config from %s", path))
	}
	secret, err := jwtSecret(config.SharedSecretFile)
	if err != nil {
		return nil, errors.Wrap(err, "error reading shared secret")
	}
	config.API.JWTSecret = secret
	config.VisualizerAPI.JWTSecret = secret
	return &config, nil
}

func jwtSecret(defaultFile string) ([]byte, error) {
	fromEnvVar := os.Getenv("JWT_SECRET")
	if fromEnvVar != "" {
		return []byte(fromEnvVar), nil
	}
	file := os.Getenv("JWT_SECRET_FILE")
	if file == "" {
		file = defaultFile
	}
	secret, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "error reading shared secret")
	}
	return bytes.TrimRight(secret, "\n"), nil
}
