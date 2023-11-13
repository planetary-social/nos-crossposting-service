package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http"
	"github.com/planetary-social/nos-crossposting-service/service/ports/http/frontend"
	"github.com/planetary-social/nos-crossposting-service/service/ports/memorypubsub"
	"github.com/planetary-social/nos-crossposting-service/service/ports/timer"
)

var portsSet = wire.NewSet(
	http.NewServer,
	http.NewMetricsServer,
	frontend.NewFrontendFileSystem,

	memorypubsub.NewReceivedEventSubscriber,
	timer.NewMetrics,
)
