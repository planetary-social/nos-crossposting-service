package apns

import (
	"sync"

	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain/notifications"
)

type APNSMock struct {
	logger logging.Logger

	sentNotificationsLock sync.Mutex
	sentNotifications     []notifications.Notification
}

func NewAPNSMock(config config.Config, logger logging.Logger) (*APNSMock, error) {
	return &APNSMock{logger: logger}, nil
}

func (a *APNSMock) SendNotification(notification notifications.Notification) error {
	a.sentNotificationsLock.Lock()
	defer a.sentNotificationsLock.Unlock()

	a.sentNotifications = append(a.sentNotifications, notification)

	a.logger.
		Debug().
		WithField("notification", notification.UUID()).
		Message("sending a mock APNs notification")

	return nil
}

func (a *APNSMock) SentNotifications() []notifications.Notification {
	a.sentNotificationsLock.Lock()
	defer a.sentNotificationsLock.Unlock()

	return internal.CopySlice(a.sentNotifications)
}
