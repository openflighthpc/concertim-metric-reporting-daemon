package domain

import "github.com/alces-flight/concertim-mrapi/config"

// Application is the application container.  It holds references to the
// various singleton components of the system such as the Repository.  It also
// has various "commands" as methods such as AddMetric.
type Application struct {
	config  config.Config
	dsmRepo DataSourceMapRepository
	Repo    Repository
}

// NewApp returns a newly configured Application.
func NewApp(config config.Config, repo Repository, dsmRepo DataSourceMapRepository) *Application {
	return &Application{Repo: repo, dsmRepo: dsmRepo}
}
