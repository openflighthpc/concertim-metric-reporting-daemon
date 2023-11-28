package rrd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

// The round robin archives (RRA) used in all of our RRD files.  The archives are:
//   - One hour consolidating 15 seconds of data.
//   - One day consolidating 5 minutes of data.
//   - 90 days consolidating of one hour of data.
//
// for each of those we consolidate the AVERAGE, MIN and MAX.
//
// These are not yet configurable as they need to be consistent with the values
// in domain.LastXLookup.
var archives = []string{
	"RRA:AVERAGE:0.5:15s:1h",
	"RRA:AVERAGE:0.5:5m:1d",
	"RRA:AVERAGE:0.5:1h:90d",
	"RRA:MIN:0.5:15s:1h",
	"RRA:MIN:0.5:5m:1d",
	"RRA:MIN:0.5:1h:90d",
	"RRA:MAX:0.5:15s:1h",
	"RRA:MAX:0.5:5m:1d",
	"RRA:MAX:0.5:1h:90d",
}

var _ domain.HistoricRepository = (*historicRepo)(nil)

type historicRepo struct {
	cluster               string
	consolidationFunction string
	dsmRepo               domain.DataSourceMapRepository
	grid                  string
	logger                zerolog.Logger
	rrdDir                string
	rrdMetricName         string
	rrdTool               string
	step                  time.Duration
}

func NewHistoricRepo(logger zerolog.Logger, config config.RRD, dsmRepo domain.DataSourceMapRepository) *historicRepo {
	return &historicRepo{
		cluster:               config.ClusterName,
		consolidationFunction: "AVERAGE",
		dsmRepo:               dsmRepo,
		grid:                  config.GridName,
		logger:                logger.With().Str("component", "historic-repo").Logger(),
		rrdDir:                config.Directory,
		rrdMetricName:         "sum",
		rrdTool:               config.ToolPath,
		step:                  config.Step,
	}
}

func (hr *historicRepo) GetValuesForHostAndMetric(
	hostId domain.HostId,
	metricName domain.MetricName,
	fetchConfig domain.HistoricMetricDuration,
) (*domain.HistoricHost, error) {
	dsm, ok := hr.dsmRepo.GetDSM(hostId)
	if !ok {
		return nil, domain.ErrHostNotFound
	}
	host := domain.HistoricHost{
		Id:      hostId,
		DSM:     dsm,
		Metrics: map[domain.MetricName][]*domain.HistoricMetric{},
	}
	cmd := fetchCmdArgs{
		clusterName: dsm.ClusterName,
		hostName:    dsm.HostName,
		metricName:  metricName,
		alignStart:  true,
		resolution:  fetchConfig.Resolution,
		startTime:   fetchConfig.Start,
		endTime:     fetchConfig.End,
	}
	metrics, err := hr.runFetchCmd(cmd)
	if err != nil {
		return nil, err
	}
	host.Metrics[metricName] = metrics
	return &host, nil
}

func (hr *historicRepo) GetValuesForMetric(
	metricName domain.MetricName,
	fetchConfig domain.HistoricMetricDuration,
) ([]*domain.HistoricHost, error) {
	hosts := make([]*domain.HistoricHost, 0)
	hostNames, err := hr.getHosts()
	if err != nil {
		return nil, fmt.Errorf("%s %w", "listing historic hosts", err)
	}
	for _, hostName := range hostNames {
		dsm := domain.DSM{
			GridName:    hr.grid,
			ClusterName: hr.cluster,
			HostName:    hostName,
		}
		hostId, ok := hr.dsmRepo.GetHostId(dsm)
		if !ok {
			hr.logger.Debug().Stringer("dsm", dsm).Msg("unknown host")
			continue
		}
		host, err := hr.GetValuesForHostAndMetric(hostId, metricName, fetchConfig)
		if err != nil {
			hr.logger.Error().Err(err).Stringer("dsm", dsm).Str("metric", string(metricName)).Msg("fetching metrics")
			continue
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (hr *historicRepo) ListMetricNames() ([]string, error) {
	path := filepath.Join(hr.rrdDir, hr.cluster, "__SummaryInfo__")
	return hr.getMetricNames(path)
}

func (hr *historicRepo) ListHostMetricNames(hostId domain.HostId) ([]string, error) {
	dsm, ok := hr.dsmRepo.GetDSM(hostId)
	if !ok {
		return nil, domain.ErrHostNotFound
	}
	path := filepath.Join(hr.rrdDir, dsm.ClusterName, dsm.HostName)
	return hr.getMetricNames(path)
}

func (hr *historicRepo) getHosts() ([]string, error) {
	cmd := exec.Command(hr.rrdTool, "list", filepath.Join(hr.rrdDir, hr.cluster))
	hr.logger.Debug().Str("cmd", cmd.String()).Msg("listing historic hosts")
	out, err := cmd.Output()
	if err != nil {
		return nil, augmentError(err, hr.rrdTool, "listing hosts")
	}
	hosts := make([]string, 0)
	for _, host := range strings.Split(string(out), "\n") {
		if host == "__SummaryInfo__" || host == "" {
			continue
		}
		hosts = append(hosts, host)
	}
	hr.logger.Debug().Strs("hosts", hosts).Msg("found hosts")
	return hosts, nil
}

func (hr *historicRepo) getMetricNames(dir string) ([]string, error) {
	cmd := exec.Command(hr.rrdTool, "list", dir)
	hr.logger.Debug().Str("cmd", cmd.String()).Msg("listing historic metrics")
	out, err := cmd.Output()
	if err != nil {
		return nil, augmentError(err, hr.rrdTool, "listing metrics")
	}
	metricNames := make([]string, 0)
	for _, file := range strings.Split(string(out), "\n") {
		if filepath.Ext(file) != ".rrd" {
			continue
		}
		metricName := strings.TrimSuffix(file, filepath.Ext(file))
		metricNames = append(metricNames, metricName)
	}
	slices.SortFunc(metricNames, strings.Compare)
	hr.logger.Debug().Strs("metricNames", metricNames).Msg("found metricNames")
	return metricNames, nil
}

func augmentError(err error, path, msg string) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if strings.Contains(exitErr.Error(), path) || strings.Contains(string(exitErr.Stderr), path) {
			return fmt.Errorf("%s: %s: %w", msg, exitErr.Stderr, exitErr)
		}
		return fmt.Errorf("%s: %s: %s: %w", msg, path, exitErr.Stderr, exitErr)
	}
	return fmt.Errorf("%s %s", msg, err)
}

type fetchCmdArgs struct {
	clusterName string
	hostName    string
	metricName  domain.MetricName
	alignStart  bool
	resolution  string
	startTime   string
	endTime     string
}

func (hr *historicRepo) runFetchCmd(args fetchCmdArgs) ([]*domain.HistoricMetric, error) {
	rrdFileName := fmt.Sprintf("%s.rrd", args.metricName)
	rrdFilePath := filepath.Join(hr.rrdDir, args.clusterName, args.hostName, rrdFileName)
	if _, err := os.Stat(rrdFilePath); errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrMetricNotFound
	}

	cmd := exec.Command(
		hr.rrdTool, "fetch", rrdFilePath, hr.consolidationFunction,
	)
	if args.alignStart {
		cmd.Args = append(cmd.Args, "--align-start")
	}
	if args.resolution != "" {
		cmd.Args = append(cmd.Args, "--resolution", args.resolution)
	}
	if args.startTime != "" {
		cmd.Args = append(cmd.Args, "--start", args.startTime)
	}
	if args.endTime != "" {
		cmd.Args = append(cmd.Args, "--end", args.endTime)
	}
	hr.logger.Debug().Str("cmd", cmd.String()).Msg("fetching metrics")
	out, err := cmd.Output()
	if err != nil {
		return nil, augmentError(err, hr.rrdTool, "fetching metrics")
	}
	hr.logger.Debug().Bytes("metrics", out).Msg("found metrics")
	return hr.parseMetricValues(out), nil
}

func (hr *historicRepo) parseMetricValues(input []byte) []*domain.HistoricMetric {
	lines := strings.Split(string(input), "\n")
	foundStart := false
	metrics := make([]*domain.HistoricMetric, 0, len(lines))
	for _, line := range lines {
		if !foundStart {
			if strings.Trim(line, " ") == hr.rrdMetricName {
				foundStart = true
			}
			continue
		}
		if line == "" {
			continue
		}
		fields := strings.Split(line, ": ")
		timestampStr := fields[0]
		valueStr := fields[1]

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			hr.logger.Error().Err(err).Msg("failed to parse timestamp")
			continue
		}
		var value float64
		if valueStr == "-nan" || valueStr == "nan" {
			value = math.NaN()
		} else {
			value, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				hr.logger.Error().Err(err).Msg("failed to parse value")
				continue
			}
		}
		metrics = append(metrics, &domain.HistoricMetric{
			Value:     value,
			Timestamp: timestamp,
		})
	}
	return metrics
}

func (hr *historicRepo) UpdateSummaryMetrics(summaries domain.MetricSummaries) error {
	var err error
	for metricName, summary := range summaries.GetSummaries() {
		hr.logger.Debug().Str("metric", string(metricName)).Int("value", summary.Num).Msg("updating consolidated metric")
		rrdFileDir := filepath.Join(hr.rrdDir, hr.cluster, "__SummaryInfo__")
		rrdFilePath := filepath.Join(rrdFileDir, fmt.Sprintf("%s.rrd", metricName))
		r := updateRunner{}
		timestamp := time.Now()
		var values string
		sumVal := reflect.ValueOf(summary.Sum)
		if sumVal.CanInt() {
			values = fmt.Sprintf("%d", sumVal.Int())
		} else if sumVal.CanUint() {
			values = fmt.Sprintf("%d", sumVal.Uint())
		} else if sumVal.CanFloat() {
			values = fmt.Sprintf("%f", sumVal.Float())
		}
		values = fmt.Sprintf("%s:%d", values, summary.Num)
		r.run(func() error { return hr.runMkdir(rrdFilePath) })
		r.run(func() error { return hr.runCreateCmd(rrdFilePath, timestamp, true) })
		r.run(func() error { return hr.runUpdateCmd(rrdFilePath, timestamp, values) })
		err = errors.Join(err, r.err)
	}
	return err
}

func (hr *historicRepo) UpdateMetric(host *domain.ProcessedHost, metric *domain.ProcessedMetric) error {
	hr.logger.Debug().Stringer("host", host.DSM).Str("metric", metric.Name).Str("value", metric.Value).Msg("updating metric")
	rrdFileDir := filepath.Join(hr.rrdDir, host.DSM.ClusterName, host.DSM.HostName)
	rrdFilePath := filepath.Join(rrdFileDir, fmt.Sprintf("%s.rrd", metric.Name))
	r := updateRunner{}
	r.run(func() error { return hr.runMkdir(rrdFilePath) })
	r.run(func() error { return hr.runCreateCmd(rrdFilePath, metric.Timestamp, false) })
	r.run(func() error { return hr.runUpdateCmd(rrdFilePath, metric.Timestamp, metric.Value) })
	return r.err
}

type updateRunner struct {
	err error
}

func (r *updateRunner) run(f func() error) {
	if r.err == nil {
		r.err = f()
	}
}

func (hr *historicRepo) runMkdir(rrdFilePath string) error {
	dirname := filepath.Dir(rrdFilePath)
	return os.MkdirAll(dirname, 0755)
}

func (hr *historicRepo) runCreateCmd(rrdFilePath string, timestamp time.Time, summary bool) error {
	if _, err := os.Stat(rrdFilePath); err == nil {
		// File already exists.
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		heartbeat := 120
		dss := []string{fmt.Sprintf("DS:sum:GAUGE:%d:NaN:NaN", heartbeat)}
		if summary {
			dss = append(dss, fmt.Sprintf("DS:num:GAUGE:%d:NaN:NaN", heartbeat))
		}
		step := int64(hr.step.Seconds())
		cmd := exec.Command(
			hr.rrdTool, "create", rrdFilePath,
			"--start", fmt.Sprintf("%d", timestamp.Unix()-step),
			"--step", fmt.Sprintf("%d", step),
			"--no-overwrite",
		)
		cmd.Args = append(cmd.Args, dss...)
		cmd.Args = append(cmd.Args, archives...)
		hr.logger.Debug().Str("cmd", cmd.String()).Msg("creating RRD file")
		out, err := cmd.Output()
		hr.logger.Debug().Str("cmd", cmd.String()).Bytes("out", out).Msg("created RRD file")
		if err != nil {
			return augmentError(err, hr.rrdTool, "creating RRD file")
		}
		return nil
	} else {
		return fmt.Errorf("%s %w", "error checking if RRD file exists", err)
	}
}

func (hr *historicRepo) runUpdateCmd(rrdFilePath string, timestamp time.Time, values string) error {
	valueSpec := fmt.Sprintf("%d:%s", timestamp.Unix(), values)
	cmd := exec.Command(
		hr.rrdTool, "update", rrdFilePath, valueSpec,
	)
	hr.logger.Debug().Str("cmd", cmd.String()).Msg("updating metrics")
	out, err := cmd.Output()
	hr.logger.Debug().Str("cmd", cmd.String()).Bytes("out", out).Msg("updated metrics")
	if err != nil {
		return augmentError(err, hr.rrdTool, "updating RRD file")
	}
	return nil
}
