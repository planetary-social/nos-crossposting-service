package timer

import (
	"context"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
)

const collectMetricsEvery = 1 * time.Minute

type Metrics struct {
	app    app.Application
	logger logging.Logger
}

func NewMetrics(app app.Application, logger logging.Logger) *Metrics {
	return &Metrics{
		app:    app,
		logger: logger.New("metrics"),
	}
}

func (m *Metrics) Run(ctx context.Context) error {
	for {
		select {
		case <-time.After(collectMetricsEvery):
			if err := m.collect(ctx); err != nil {
				m.logger.Error().WithError(err).Message("error triggering app handler")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Metrics) collect(ctx context.Context) error {
	return m.app.UpdateMetrics.Handle(ctx)
}
