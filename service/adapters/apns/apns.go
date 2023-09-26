package apns

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type Metrics interface {
	ReportCallToAPNS(statusCode int, err error)
}

type APNS struct {
	client  *apns2.Client
	cfg     config.Config
	metrics Metrics
	logger  logging.Logger
}

func NewAPNS(cfg config.Config, metrics Metrics, logger logging.Logger) (*APNS, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error creating an apns client")
	}
	return &APNS{
		client:  client,
		cfg:     cfg,
		metrics: metrics,
		logger:  logger.New("apns"),
	}, nil
}

func newClient(cfg config.Config) (*apns2.Client, error) {
	cert, err := certificate.FromP12File(cfg.APNSCertificatePath(), cfg.APNSCertificatePassword())
	if err != nil {
		return nil, errors.Wrap(err, "error loading certificate")
	}

	switch cfg.Environment() {
	case config.EnvironmentProduction:
		return apns2.NewClient(cert).Production(), nil
	case config.EnvironmentDevelopment:
		return apns2.NewClient(cert).Development(), nil
	default:
		return nil, errors.New("unknown environment")
	}
}

func (a *APNS) SendNotification(notification notifications.Notification) error {
	n := &apns2.Notification{}
	n.PushType = apns2.PushTypeBackground
	n.ApnsID = notification.UUID().String()
	n.DeviceToken = notification.APNSToken().Hex()
	n.Topic = a.cfg.APNSTopic()
	n.Payload = notification.Payload()
	n.Priority = apns2.PriorityLow

	resp, err := a.client.Push(n)
	a.metrics.ReportCallToAPNS(resp.StatusCode, err)
	if err != nil {
		return errors.Wrap(err, "error pushing the notification")
	}

	a.logger.Debug().
		WithField("uuid", notification.UUID().String()).
		WithField("response.reason", resp.Reason).
		WithField("response.statusCode", resp.StatusCode).
		WithField("host", a.client.Host).
		Message("sent a notification")

	return nil
}
