package rrd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

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
	}
}

func (hr *historicRepo) GetValuesForHostAndMetric(
	hostId domain.HostId,
	metricName domain.MetricName,
	startTime time.Time,
	endTime time.Time,
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
	metrics, err := hr.getMetricValues(dsm, metricName, startTime, endTime)
	if err != nil {
		return nil, err
	}
	host.Metrics[metricName] = metrics
	return &host, nil
}

func (hr *historicRepo) GetValuesForMetric(metricName domain.MetricName, startTime, endTime time.Time) ([]*domain.HistoricHost, error) {
	hosts := make([]*domain.HistoricHost, 0)
	hostNames, err := hr.getHosts()
	if err != nil {
		return nil, errors.Wrap(err, "listing historic hosts")
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
		host := domain.HistoricHost{
			Id:      hostId,
			DSM:     dsm,
			Metrics: map[domain.MetricName][]*domain.HistoricMetric{},
		}
		metrics, err := hr.getMetricValues(dsm, metricName, startTime, endTime)
		if err != nil {
			hr.logger.Error().Err(err).Stringer("dsm", dsm).Str("metric", string(metricName)).Msg("fetching metrics")
		}
		host.Metrics[metricName] = metrics
		hosts = append(hosts, &host)
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
		return nil, augmentError(err, hr.rrdTool, "executing")
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
		return nil, augmentError(err, hr.rrdTool, "executing")
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

func (hr *historicRepo) getMetricValues(
	dsm domain.DSM,
	metricName domain.MetricName,
	startTime, endTime time.Time,
) ([]*domain.HistoricMetric, error) {
	rrdFileName := fmt.Sprintf("%s.rrd", metricName)
	rrdFilePath := filepath.Join(hr.rrdDir, dsm.ClusterName, dsm.HostName, rrdFileName)
	if _, err := os.Stat(rrdFilePath); errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrMetricNotFound
	}
	cmd := exec.Command(
		hr.rrdTool, "fetch", rrdFilePath, hr.consolidationFunction,
		"-s", fmt.Sprintf("%d", startTime.Unix()),
		"-e", fmt.Sprintf("%d", endTime.Unix()),
	)
	hr.logger.Debug().Str("cmd", cmd.String()).Msg("fetching metrics")
	out, err := cmd.Output()
	if err != nil {
		return nil, augmentError(err, hr.rrdTool, "executing")
	}
	hr.logger.Debug().Bytes("metrics", out).Msg("found metrics")
	return hr.parseMetricValues(out), nil
}

func augmentError(err error, path, msg string) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if strings.Contains(exitErr.Error(), path) || strings.Contains(string(exitErr.Stderr), path) {
			return errors.Wrapf(exitErr, "%s: %s", msg, exitErr.Stderr)
		}
		return errors.Wrapf(exitErr, "%s: %s: %s", msg, path, exitErr.Stderr)
	}
	return errors.Wrap(err, msg)
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
