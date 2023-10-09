package domain

import "github.com/alces-flight/concertim-metric-reporting-daemon/config"

// Application is the application container.  It holds references to the
// various singleton components of the system such as the Repository.  It also
// has various "commands" as methods such as AddMetric.
type Application struct {
	Repo         ReportedRepository
	config       config.Config
	dsmRepo      DataSourceMapRepository
	dsmUpdater   DataSourceMapRepoUpdater
	ResultRepo   ProcessedRepository
	HistoricRepo HistoricRepository
}

// NewApp returns a newly configured Application.
func NewApp(
	config config.Config,
	repo ReportedRepository,
	dsmRepo DataSourceMapRepository,
	dsmUpdater DataSourceMapRepoUpdater,
	resultRepo ProcessedRepository,
	historicRepo HistoricRepository,
) *Application {
	return &Application{
		Repo:         repo,
		config:       config,
		dsmRepo:      dsmRepo,
		dsmUpdater:   dsmUpdater,
		ResultRepo:   resultRepo,
		HistoricRepo: historicRepo,
	}
}
