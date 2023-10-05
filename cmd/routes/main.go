package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/alces-flight/concertim-metric-reporting-daemon/api"
	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

var configFile = flag.String("config-file", config.DefaultPath, "path to config file")

func init() {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	isatty := err == nil
	if isatty {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
}

func setLogLevel(config *config.Config) {
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Error().Err(err).Msg("Unable to set log level")
	}
	zerolog.SetGlobalLevel(level)
}

func loadConfig() (*config.Config, error) {
	if *configFile == "" {
		return config.FromFile(config.DefaultPath)
	} else {
		return config.FromFile(*configFile)
	}
}

func getFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	strs := strings.Split((name), "/")
	return strs[len(strs)-1]
}

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
		os.Exit(1)
	}
	config, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("loading config failed")
	}
	setLogLevel(config)

	app := domain.NewApp(*config, nil, nil, nil, nil, nil)
	apiServer := api.NewServer(log.Logger, app, config.API)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		fmt.Printf("%s %s\t%s\n", method, route, getFunctionName(handler))
		return nil
	}

	if err := chi.Walk(apiServer.Router, walkFunc); err != nil {
		fmt.Printf("Logging err: %s\n", err.Error())
	}
}
