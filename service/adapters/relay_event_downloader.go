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
	reconnectAfter           = 1 * time.Minute
	manageSubscriptionsEvery = 10 * time.Second
)

type RelayEventDownloader struct {
	logger logging.Logger

	ctx context.Context

	connections     map[domain.RelayAddress]*RelayConnection
	connectionsLock sync.Mutex
}

func NewRelayEventDownloader(ctx context.Context, logger logging.Logger) *RelayEventDownloader {
	return &RelayEventDownloader{
		logger:          logger.New("relayEventDownloader"),
		ctx:             ctx,
		connections:     make(map[domain.RelayAddress]*RelayConnection),
		connectionsLock: sync.Mutex{},
	}
}

func (r *RelayEventDownloader) GetEvents(ctx context.Context, publicKey domain.PublicKey, relayAddress domain.RelayAddress, eventKinds []domain.EventKind, maxAge *time.Duration) <-chan app.EventOrEndOfSavedEvents {
	connection := r.getConnection(relayAddress)
	return connection.GetEvents(ctx, publicKey, eventKinds, maxAge)
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
