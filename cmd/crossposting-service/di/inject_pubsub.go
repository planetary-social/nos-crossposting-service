package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	sqlitepubsubport "github.com/planetary-social/nos-crossposting-service/service/ports/sqlitepubsub"
)

var memoryPubsubSet = wire.NewSet(
	memorypubsub.NewReceivedEventPubSub,
	wire.Bind(new(app.ReceivedEventPublisher), new(*memorypubsub.ReceivedEventPubSub)),
	wire.Bind(new(app.ReceivedEventSubscriber), new(*memorypubsub.ReceivedEventPubSub)),
)

var sqlitePubsubSet = wire.NewSet(
	sqlite.NewWatermillSchemaAdapter,
	sqlite.NewWatermillOffsetsAdapter,
	sqlite.NewWatermillSubscriber,
	sqlitepubsubport.NewTweetCreatedEventSubscriber,
)

var sqliteTxPubsubSet = wire.NewSet(
	sqlite.NewWatermillPublisher,
	sqlite.NewPublisher,
	wire.Bind(new(app.Publisher), new(*sqlite.Publisher)),
)
