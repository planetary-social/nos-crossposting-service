package di

import (
	"context"

	googlefirestore "cloud.google.com/go/firestore"
	watermillfirestore "github.com/ThreeDotsLabs/watermill-firestore/pkg/firestore"
	"github.com/boreq/errors"
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/apns"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/firestore"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/ports/firestorepubsub"
)

var firestoreAdaptersSet = wire.NewSet(
	newFirestoreClient,

	firestore.NewTransactionProvider,
	wire.Bind(new(app.TransactionProvider), new(*firestore.TransactionProvider)),

	newAdaptersFactoryFn,

	firestore.NewWatermillSubscriber,
	wire.Bind(new(firestorepubsub.FirestoreSubscriber), new(*watermillfirestore.Subscriber)),

	wire.Struct(new(buildTransactionFirestoreAdaptersDependencies), "*"),
)

func newAdaptersFactoryFn(deps buildTransactionFirestoreAdaptersDependencies) firestore.AdaptersFactoryFn {
	return func(client *googlefirestore.Client, tx *googlefirestore.Transaction) (app.Adapters, error) {
		return buildTransactionFirestoreAdapters(client, tx, deps)
	}
}

var firestoreTxAdaptersSet = wire.NewSet(
	firestore.NewRegistrationRepository,
	wire.Bind(new(app.RegistrationRepository), new(*firestore.RegistrationRepository)),

	firestore.NewEventRepository,
	wire.Bind(new(app.EventRepository), new(*firestore.EventRepository)),

	firestore.NewRelayRepository,
	wire.Bind(new(app.RelayRepository), new(*firestore.RelayRepository)),

	firestore.NewPublicKeyRepository,
	wire.Bind(new(app.PublicKeyRepository), new(*firestore.PublicKeyRepository)),

	firestore.NewTagRepository,
	wire.Bind(new(app.TagRepository), new(*firestore.TagRepository)),

	firestore.NewWatermillPublisher,
	firestore.NewPublisher,
	wire.Bind(new(app.Publisher), new(*firestore.Publisher)),
)

var adaptersSet = wire.NewSet(
	apns.NewAPNS,
	wire.Bind(new(app.APNS), new(*apns.APNS)),

	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(firestorepubsub.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(apns.Metrics), new(*prometheus.Prometheus)),

	adapters.NewMemoryEventWasAlreadySavedCache,
	wire.Bind(new(app.EventWasAlreadySavedCache), new(*adapters.MemoryEventWasAlreadySavedCache)),
)

var integrationAdaptersSet = wire.NewSet(
	apns.NewAPNSMock,
	wire.Bind(new(app.APNS), new(*apns.APNSMock)),

	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(firestorepubsub.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(apns.Metrics), new(*prometheus.Prometheus)),

	adapters.NewMemoryEventWasAlreadySavedCache,
	wire.Bind(new(app.EventWasAlreadySavedCache), new(*adapters.MemoryEventWasAlreadySavedCache)),
)

func newFirestoreClient(ctx context.Context, config config.Config, logger logging.Logger) (*googlefirestore.Client, func(), error) {
	v, err := firestore.NewClient(ctx, config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating the firestore client")
	}

	return v, func() {
		if err := v.Close(); err != nil {
			logger.Error().WithError(err).Message("error closing firestore")
		}
	}, nil
}
