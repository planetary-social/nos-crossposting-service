package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/pubsub"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(app.Application), "*"),

	app.NewProcessReceivedEventHandler,
	wire.Bind(new(memorypubsub.SaveReceivedEventHandler), new(*app.ProcessReceivedEventHandler)),

	app.NewSendTweetHandler,
	wire.Bind(new(pubsub.SendTweetHandler), new(*app.SendTweetHandler)),

	app.NewGetSessionAccountHandler,
	app.NewLoginOrRegisterHandler,
	app.NewLinkPublicKeyHandler,
)
