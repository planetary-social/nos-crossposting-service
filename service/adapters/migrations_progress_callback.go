package adapters

import "github.com/planetary-social/nos-crossposting-service/internal/logging"

type LoggingMigrationsProgressCallback struct {
	logger logging.Logger
}

func NewLoggingMigrationsProgressCallback(logger logging.Logger) *LoggingMigrationsProgressCallback {
	return &LoggingMigrationsProgressCallback{logger: logger.New("migrationProgressCallback")}
}

func (l LoggingMigrationsProgressCallback) OnRunning(migrationIndex int, migrationsCount int) {
	l.logger.Debug().
		WithField("index", migrationIndex).
		WithField("count", migrationsCount).
		Message("running")
}

func (l LoggingMigrationsProgressCallback) OnError(migrationIndex int, migrationsCount int, err error) {
	l.logger.Error().
		WithField("index", migrationIndex).
		WithField("count", migrationsCount).
		WithError(err).
		Message("error")
}

func (l LoggingMigrationsProgressCallback) OnDone(migrationsCount int) {
	l.logger.Debug().
		WithField("count", migrationsCount).
		Message("done")
}
