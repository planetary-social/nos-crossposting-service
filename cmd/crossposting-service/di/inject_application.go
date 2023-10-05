package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
)

var applicationSet = wire.NewSet(
	wire.Struct(new(app.Application), "*"),

	app.NewSaveReceivedEventHandler,
	wire.Bind(new(memorypubsub.SaveReceivedEventHandler), new(*app.SaveReceivedEventHandler)),

	app.NewGetSessionAccountHandler,
	app.NewLoginOrRegisterHandler,
	app.NewLinkPublicKeyHandler,
)
