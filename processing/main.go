package main

import (
	"os"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/retrieval"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

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

func main() {
	config, err := config.FromFile(config.DefaultPaths)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}
	setLogLevel(config)
	pollChan := make(chan []retrieval.Grid)
	poller, err := retrieval.New(log.Logger, config.Gmetad)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create retrieval.poller")
	}

	go func() {
		err = poller.Start(pollChan)
		if err != nil {
			log.Fatal().Err(err).Msg("Retrieving metrics")
		}
	}()

	for grids := range pollChan {
		log.Info().Int("len", len(grids)).Msg("Got grids")
		for _, grid := range grids {
			for _, cluster := range grid.Clusters {
				for _, host := range cluster.Hosts {
					log.Info().
						Str("grid", grid.Name).
						Str("cluster", cluster.Name).
						Str("host", host.Name).
						Int("metric.count", len(host.Metrics)).
						Msg("got host")
				}
			}
		}

	}
}
