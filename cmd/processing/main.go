// Package main runs the ganglia metric retriever, the metric processor and the
// metric reporter.
package main

import (
	"flag"
	"os"

	"github.com/alces-flight/concertim-metric-reporting-daemon/dsmRepository"
	"github.com/alces-flight/concertim-metric-reporting-daemon/processing"
	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/retrieval"
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

func main() {
	flag.Parse()
	config, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}
	setLogLevel(config)
	pollChan := make(chan []retrieval.Grid)
	poller, err := retrieval.New(log.Logger, config.Retrieval)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create retrieval.poller")
	}

	dsmRepo := dsmRepository.New(log.Logger, config.DSM)
	processor := processing.NewProcessor(log.Logger, dsmRepo)
	recorder := processing.NewScriptRecorder(log.Logger, config.Recorder)

	go func() {
		err = poller.Start(pollChan)
		if err != nil {
			log.Fatal().Err(err).Msg("Retrieving metrics")
		}
	}()

	for grids := range pollChan {
		results, err := processor.Process(grids)
		if err != nil {
			log.Err(err).Msg("processing metrics")
		}
		err = recorder.Record(results)
		if err != nil {
			log.Err(err).Msg("recording results")
		}
	}
}
