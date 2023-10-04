package di

import (
	googlefirestore "cloud.google.com/go/firestore"
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/firestore"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/memory"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/prometheus"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/firestorepubsub"
)

var firestoreAdaptersSet = wire.NewSet(
	//newFirestoreClient,

	firestore.NewTransactionProvider,
	wire.Bind(new(app.TransactionProvider), new(*firestore.TransactionProvider)),

	newAdaptersFactoryFn,

	//firestore.NewWatermillSubscriber,
	//wire.Bind(new(firestorepubsub.FirestoreSubscriber), new(*watermillfirestore.Subscriber)),

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

	//firestore.NewWatermillPublisher,
	firestore.NewPublisher,
	//wire.Bind(new(app.Publisher), new(*firestore.Publisher)),

	memory.NewMemoryAccountRepository,
	wire.Bind(new(app.AccountRepository), new(*memory.MemoryAccountRepository)),

	memory.NewMemorySessionRepository,
	wire.Bind(new(app.SessionRepository), new(*memory.MemorySessionRepository)),
)

var adaptersSet = wire.NewSet(
	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(firestorepubsub.Metrics), new(*prometheus.Prometheus)),

	adapters.NewMemoryEventWasAlreadySavedCache,
	wire.Bind(new(app.EventWasAlreadySavedCache), new(*adapters.MemoryEventWasAlreadySavedCache)),

	adapters.NewIDGenerator,
	wire.Bind(new(app.SessionIDGenerator), new(*adapters.IDGenerator)),
	wire.Bind(new(app.AccountIDGenerator), new(*adapters.IDGenerator)),

	memory.NewMemoryAccountRepository,
	wire.Bind(new(app.AccountRepository), new(*memory.MemoryAccountRepository)),

	memory.NewMemorySessionRepository,
	wire.Bind(new(app.SessionRepository), new(*memory.MemorySessionRepository)),
)

var integrationAdaptersSet = wire.NewSet(
	prometheus.NewPrometheus,
	wire.Bind(new(app.Metrics), new(*prometheus.Prometheus)),
	wire.Bind(new(firestorepubsub.Metrics), new(*prometheus.Prometheus)),

	adapters.NewMemoryEventWasAlreadySavedCache,
	wire.Bind(new(app.EventWasAlreadySavedCache), new(*adapters.MemoryEventWasAlreadySavedCache)),

	adapters.NewIDGenerator,
	wire.Bind(new(app.SessionIDGenerator), new(*adapters.IDGenerator)),
	wire.Bind(new(app.AccountIDGenerator), new(*adapters.IDGenerator)),

	memory.NewMemoryAccountRepository,
	wire.Bind(new(app.AccountRepository), new(*memory.MemoryAccountRepository)),

	memory.NewMemorySessionRepository,
	wire.Bind(new(app.SessionRepository), new(*memory.MemorySessionRepository)),
)

//func newFirestoreClient(ctx context.Context, config config.Config, logger logging.Logger) (*googlefirestore.Client, func(), error) {
//	v, err := firestore.NewClient(ctx, config)
//	if err != nil {
//		return nil, nil, errors.Wrap(err, "error creating the firestore client")
//	}
//
//	return v, func() {
//		if err := v.Close(); err != nil {
//			logger.Error().WithError(err).Message("error closing firestore")
//		}
//	}, nil
//}
