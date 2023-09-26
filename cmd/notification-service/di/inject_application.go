package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/firestorepubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(app.Application), "*"),

	commandsSet,
	queriesSet,
)

var commandsSet = wire.NewSet(
	wire.Struct(new(app.Commands), "*"),

	app.NewSaveRegistrationHandler,

	app.NewSaveReceivedEventHandler,
	wire.Bind(new(memorypubsub.SaveReceivedEventHandler), new(*app.SaveReceivedEventHandler)),

	app.NewProcessSavedEventHandler,
	wire.Bind(new(firestorepubsub.ProcessSavedEventHandler), new(*app.ProcessSavedEventHandler)),
)

var queriesSet = wire.NewSet(
	wire.Struct(new(app.Queries), "*"),

	app.NewGetRelaysHandler,
	app.NewGetPublicKeysHandler,
	app.NewGetTokensHandler,
	app.NewGetEventsHandler,
	app.NewGetNotificationsHandler,
)
