// Package main runs the HTTP API server and the processing loop.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"

	"github.com/alces-flight/concertim-metric-reporting-daemon/api"
	"github.com/alces-flight/concertim-metric-reporting-daemon/canned"
	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/dsmRepository"
	"github.com/alces-flight/concertim-metric-reporting-daemon/inmem"
	"github.com/alces-flight/concertim-metric-reporting-daemon/rrd"
	"github.com/alces-flight/concertim-metric-reporting-daemon/visualizer"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write mem profile to file")
	configFile = flag.String("config-file", config.DefaultPath, "path to config file")
)

var Usage = func() {
	cmd := path.Base(os.Args[0])
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage: %s [OPTION]... [COMMAND]\n", cmd)
	fmt.Fprintf(w, "\nThe commands are:\n\n")
	fmt.Fprintf(w, "\tversion\t print version\n")
	fmt.Fprintf(w, "\nThe options are:\n\n")
	flag.PrintDefaults()
}

var version string

func init() {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	isatty := err == nil
	if isatty {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
	flag.Usage = Usage
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
	if flag.Arg(0) == "version" {
		cmd := path.Base(os.Args[0])
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s version %s\n", cmd, version)
		os.Exit(0)
	} else if len(flag.Args()) > 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		defer pprof.StopCPUProfile()
	}
	config, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("loading config failed")
	}
	setLogLevel(config)
	pendingRepo := inmem.NewPendingRepository(log.Logger)
	dsmRetriever := getDSMRetriever(config)
	dsmRepo := inmem.NewDSMRepo(log.Logger, config.DSM)
	dsmUpdater := dsmRepository.NewUpdater(log.Logger, config.DSM, dsmRepo, dsmRetriever)
	currentRepo := inmem.NewProcessedRepository(log.Logger)
	historicRepo := rrd.NewHistoricRepo(log.Logger, config.RRD, dsmRepo)
	app := domain.NewApp(pendingRepo, dsmRepo, dsmUpdater, currentRepo, historicRepo)
	apiServer := api.NewServer(log.Logger, app, config.API)
	go func() {
		err := apiServer.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("api.Server closed. Waiting for active connections to finish")
		} else if err != nil {
			log.Fatal().Err(err).Msg("api.Server.ListenAndServe")
		}
	}()
	go func() {
		runMetricProcessor(config, pendingRepo, currentRepo, historicRepo)
	}()

	gracefulExitSigs := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, gracefulExitSigs...)
	<-sigint
	log.Info().Msg("Closing connections")
	signal.Reset(gracefulExitSigs...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("http.Server.Shutdown")
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		f.Close() //nolint:errcheck
		return
	}
}

func runMetricProcessor(
	config *config.Config,
	pendingRepo domain.PendingRepository,
	currentRepo domain.CurrentRepository,
	historicRepo domain.HistoricRepository,
) {
	step := config.RRD.Step
	processor := domain.NewProcessor(pendingRepo, currentRepo, historicRepo, step, log.Logger)
	ticker := time.NewTicker(step)
	for {
		<-ticker.C
		processor.Process()
	}
}

func getDSMRetriever(config *config.Config) domain.DataSourceMapRetreiver {
	if config.DSM.Testdata != "" {
		return &canned.DSMRetriever{
			Path:   config.DSM.Testdata,
			Logger: log.Logger,
		}
	}
	return visualizer.New(log.Logger, config.VisualizerAPI)
}
