// Package main runs the HTTP API server and the TCP GDS server.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
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
	"github.com/alces-flight/concertim-metric-reporting-daemon/gds"
	"github.com/alces-flight/concertim-metric-reporting-daemon/inmem"
	"github.com/alces-flight/concertim-metric-reporting-daemon/processing"
	"github.com/alces-flight/concertim-metric-reporting-daemon/repository/memory"
	"github.com/alces-flight/concertim-metric-reporting-daemon/retrieval"
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
	repository := memory.New(log.Logger)
	dsmRetriever := getDSMRetriever(config)
	dsmRepo := inmem.NewDSMRepo(log.Logger, config.DSM)
	dsmUpdater := dsmRepository.NewUpdater(log.Logger, config.DSM, dsmRepo, dsmRetriever)
	processedRepo := inmem.NewProcessedRepository(log.Logger)
	historicRepo := rrd.NewHistoricRepo(log.Logger, config.RRD, dsmRepo)
	app := domain.NewApp(*config, repository, dsmRepo, dsmUpdater, processedRepo, historicRepo)
	apiServer := api.NewServer(log.Logger, app, config.API)
	gdsServer, err := gds.New(log.Logger, app, config.GDS)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create gds.Server")
	}
	go func() {
		err := apiServer.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("api.Server closed. Waiting for active connections to finish")
		} else if err != nil {
			log.Fatal().Err(err).Msg("api.Server.ListenAndServe")
		}
	}()
	go func() {
		err := gdsServer.ListenAndServe()
		if err != nil && errors.Is(err, net.ErrClosed) {
			log.Info().Msg("gds.Server closed")
		} else if err != nil {
			log.Fatal().Err(err).Msg("gds.Server.ListenAndServe")
		}
	}()
	go func() {
		err := runMetricProcessor(config, dsmRepo, dsmUpdater, gdsServer, processedRepo)
		if err != nil {
			log.Fatal().Err(err).Msg("running metric processor")
		}
	}()

	gracefulExitSigs := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, gracefulExitSigs...)
	<-sigint
	log.Info().Msg("Closing connections")
	signal.Reset(gracefulExitSigs...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := gdsServer.Close(); err != nil {
		log.Error().Err(err).Msg("gds.Server.Close")
	}
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
	dsmRepo domain.DataSourceMapRepository,
	dsmUpdater domain.DataSourceMapRepoUpdater,
	gdsServer *gds.Server,
	resultRepo domain.ProcessedRepository,
) error {
	pollChan := make(chan []*domain.ProcessedHost)
	poller, err := retrieval.New(log.Logger, config.Retrieval, dsmRepo, dsmUpdater)
	if err != nil {
		return errors.Wrap(err, "creating retrieval poller")
	}
	processor := processing.NewProcessor(resultRepo, log.Logger)

	// Start the ganglia metric poller.  It will report polled hosts on pollChan.
	go func() { poller.Start(pollChan) }()

	// Each time we report metrics to gmetad, kick the processing loop.
	go func() {
		for {
			<-gdsServer.AcceptedChan
			time.Sleep(config.Retrieval.PostGmetadDelay)
			poller.Ticker.TickNow()
		}
	}()

	// Each time we poll metrics from ganglia process them.
	for hosts := range pollChan {
		err := processor.Process(hosts)
		if err != nil {
			log.Error().Err(err).Msg("processing metrics")
		}
	}
	return nil
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
