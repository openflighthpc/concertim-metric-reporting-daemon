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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/openflighthpc/concertim-metric-reporting-daemon/config"
	"github.com/go-chi/jwtauth/v5"
)

var configFile = flag.String("config-file", config.DefaultPath, "path to config file")

func loadConfig() (*config.Config, error) {
	if *configFile == "" {
		return config.FromFile(config.DefaultPath)
	} else {
		return config.FromFile(*configFile)
	}
}

func main() {
	flag.Parse()
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
	tokenAuth := jwtauth.New("HS256", config.API.JWTSecret, nil)
	expiresIn := time.Hour * 24

	claims := map[string]interface{}{}
	jwtauth.SetExpiryIn(claims, expiresIn)

	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create token: %s\n", err.Error())
	}

	fmt.Print(tokenString, "\n")
}
