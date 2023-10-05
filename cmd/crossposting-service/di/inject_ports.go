package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
)

var portsSet = wire.NewSet(
	http.NewServer,
	http.NewMetricsServer,

	memorypubsub.NewReceivedEventSubscriber,
	//firestorepubsub.NewEventSavedSubscriber,
)
