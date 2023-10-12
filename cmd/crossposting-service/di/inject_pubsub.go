package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/pubsub"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	pubsubport "github.com/planetary-social/nos-crossposting-service/service/ports/pubsub"
)

var memoryPubsubSet = wire.NewSet(
	pubsub.NewReceivedEventPubSub,
	wire.Bind(new(app.ReceivedEventPublisher), new(*pubsub.ReceivedEventPubSub)),
	wire.Bind(new(app.ReceivedEventSubscriber), new(*pubsub.ReceivedEventPubSub)),
)

var sqlitePubsubSet = wire.NewSet(
	sqlite.NewWatermillSchemaAdapter,
	sqlite.NewWatermillOffsetsAdapter,
	sqlite.NewWatermillSubscriber,
	pubsubport.NewTweetCreatedEventSubscriber,
)

var sqliteTxPubsubSet = wire.NewSet(
	sqlite.NewWatermillPublisher,
	sqlite.NewPublisher,
	wire.Bind(new(app.Publisher), new(*sqlite.Publisher)),
)
