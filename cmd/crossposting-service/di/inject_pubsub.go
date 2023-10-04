package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/pubsub"
	"github.com/planetary-social/nos-crossposting-service/service/app"
)

var pubsubSet = wire.NewSet(
	pubsub.NewReceivedEventPubSub,
	wire.Bind(new(app.ReceivedEventPublisher), new(*pubsub.ReceivedEventPubSub)),
	wire.Bind(new(app.ReceivedEventSubscriber), new(*pubsub.ReceivedEventPubSub)),
)
