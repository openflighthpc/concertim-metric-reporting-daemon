package domain

import "github.com/alces-flight/concertim-metric-reporting-daemon/config"

// Application is the application container.  It holds references to the
// various singleton components of the system such as the Repository.  It also
// has various "commands" as methods such as AddMetric.
type Application struct {
	Repo       Repository
	config     config.Config
	dsmRepo    DataSourceMapRepository
	ResultRepo ResultRepo
}

// NewApp returns a newly configured Application.
func NewApp(config config.Config, repo Repository, dsmRepo DataSourceMapRepository, resultRepo ResultRepo) *Application {
	return &Application{Repo: repo, dsmRepo: dsmRepo, ResultRepo: resultRepo}
}
