package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

const (
	storeMetricsEvery = 30 * time.Second

	reconnectAfter           = 1 * time.Minute
	manageSubscriptionsEvery = 10 * time.Second
)

type RelayEventDownloader struct {
	logger  logging.Logger
	metrics app.Metrics

	ctx context.Context

	connections     map[domain.RelayAddress]*RelayConnection
	connectionsLock sync.Mutex
}

func NewRelayEventDownloader(ctx context.Context, logger logging.Logger, metrics app.Metrics) *RelayEventDownloader {
	v := &RelayEventDownloader{
		logger:          logger.New("relayEventDownloader"),
		metrics:         metrics,
		ctx:             ctx,
		connections:     make(map[domain.RelayAddress]*RelayConnection),
		connectionsLock: sync.Mutex{},
	}
	go v.storeMetricsLoop(ctx)
	return v
}

func (r *RelayEventDownloader) GetEvents(ctx context.Context, publicKey domain.PublicKey, relayAddress domain.RelayAddress, eventKinds []domain.EventKind, maxAge *time.Duration) <-chan app.EventOrEndOfSavedEvents {
	connection := r.getConnection(relayAddress)
	return connection.GetEvents(ctx, publicKey, eventKinds, maxAge)
}

func (d *RelayEventDownloader) storeMetricsLoop(ctx context.Context) {
	for {
		d.storeMetrics()

		select {
		case <-time.After(storeMetricsEvery):
		case <-ctx.Done():
			return
		}
	}
}

func (d *RelayEventDownloader) storeMetrics() {
	d.connectionsLock.Lock()
	defer d.connectionsLock.Unlock()

	for _, connection := range d.connections {
		d.metrics.ReportRelayConnectionState(connection.Address(), connection.State())
	}
}

func (r *RelayEventDownloader) getConnection(relayAddress domain.RelayAddress) *RelayConnection {
	r.connectionsLock.Lock()
	defer r.connectionsLock.Unlock()

	if connection, ok := r.connections[relayAddress]; ok {
		return connection
	}

	connection := NewRelayConnection(relayAddress, r.logger)
	go connection.Run(r.ctx)

	r.connections[relayAddress] = connection
	return connection
}
