package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
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
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err.Error())
		os.Exit(1)
	}
	tokenAuth := jwtauth.New("HS256", config.JWTSecret, nil)
	expiresIn := time.Hour * 24

	claims := map[string]interface{}{}
	jwtauth.SetExpiryIn(claims, expiresIn)

	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create token: %s\n", err.Error())
	}

	fmt.Print(tokenString, "\n")
}
