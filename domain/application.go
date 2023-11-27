package domain

// Application is the application container.  It holds references to the
// various singleton components of the system such as the Repository.  It also
// has various "commands" as methods such as AddMetric.
type Application struct {
	Repo         PendingRepository
	dsmRepo      DataSourceMapRepository
	dsmUpdater   DataSourceMapRepoUpdater
	CurrentRepoo CurrentRepository
	HistoricRepo HistoricRepository
}

// NewApp returns a newly configured Application.
func NewApp(
	repo PendingRepository,
	dsmRepo DataSourceMapRepository,
	dsmUpdater DataSourceMapRepoUpdater,
	resultRepo CurrentRepository,
	historicRepo HistoricRepository,
) *Application {
	return &Application{
		Repo:         repo,
		dsmRepo:      dsmRepo,
		dsmUpdater:   dsmUpdater,
		CurrentRepoo: resultRepo,
		HistoricRepo: historicRepo,
	}
}
