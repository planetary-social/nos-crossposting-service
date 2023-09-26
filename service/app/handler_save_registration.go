package app

import (
	"context"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type SaveRegistration struct {
	registration domain.Registration
}

func NewSaveRegistration(registration domain.Registration) SaveRegistration {
	return SaveRegistration{registration: registration}
}

type SaveRegistrationHandler struct {
	transactionProvider TransactionProvider
	logger              logging.Logger
	metrics             Metrics
}

func NewSaveRegistrationHandler(
	transactionProvider TransactionProvider,
	logger logging.Logger,
	metrics Metrics,
) *SaveRegistrationHandler {
	return &SaveRegistrationHandler{
		transactionProvider: transactionProvider,
		logger:              logger.New("saveRegistrationHandler"),
		metrics:             metrics,
	}
}

func (h *SaveRegistrationHandler) Handle(ctx context.Context, cmd SaveRegistration) (err error) {
	defer h.metrics.StartApplicationCall("saveRegistration").End(&err)

	h.logger.Debug().
		WithField("apnsToken", cmd.registration.APNSToken().Hex()).
		WithField("publicKey", cmd.registration.PublicKey().Hex()).
		WithField("relays", cmd.registration.Relays()).
		Message("saving registration")

	return h.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		return adapters.Registrations.Save(cmd.registration)
	})
}
