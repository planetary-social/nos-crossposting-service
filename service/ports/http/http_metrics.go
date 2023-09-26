package http

import (
	"context"
	"net"
	"net/http"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	prometheusadapters "github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	prometheus *prometheusadapters.Prometheus
	config     config.Config
	logger     logging.Logger
}

func NewMetricsServer(
	prometheus *prometheusadapters.Prometheus,
	config config.Config,
	logger logging.Logger,
) MetricsServer {
	return MetricsServer{
		prometheus: prometheus,
		config:     config,
		logger:     logger.New("metricsServer"),
	}
}

func (s *MetricsServer) ListenAndServe(ctx context.Context) error {
	mux := s.createMux()

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", s.config.MetricsListenAddress())
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			s.logger.Error().WithError(err).Message("error closing listener")
		}
	}()

	return http.Serve(listener, mux)
}

func (s *MetricsServer) createMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(s.prometheus.Registry(), promhttp.HandlerOpts{}))
	return mux
}
