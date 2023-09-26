package main

import (
	"context"
	"fmt"
	"github.com/boreq/errors"
	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/cmd/notification-service/di"
	configadapters "github.com/planetary-social/nos-crossposting-service/service/adapters/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	if len(os.Args) != 2 {
		return errors.New("usage: program <npub>")
	}

	publicKey, err := domain.NewPublicKeyFromNpub(os.Args[1])
	if err != nil {
		return errors.Wrap(err, "error decoding the npub")
	}

	cfg, err := configadapters.NewEnvironmentConfigLoader().Load()
	if err != nil {
		return errors.Wrap(err, "error creating a config")
	}

	service, cleanup, err := di.BuildService(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "error building a service")
	}
	defer cleanup()

	filters, err := domain.NewFilters(nostr.Filters{
		{
			Authors: []string{
				publicKey.Hex(),
			},
			Limit: 100,
		},
	})
	if err != nil {
		return errors.Wrap(err, "error creating filters")
	}

	for v := range service.App().Queries.GetEvents.Handle(ctx, filters) {
		if err := v.Err(); err != nil {
			return errors.Wrap(err, "handler returned an error")
		}

		if v.EOSE() {
			break
		}

		evt := v.Event()

		notifications, err := service.App().Queries.GetNotifications.Handle(ctx, v.Event().Id())
		if err != nil {
			return errors.Wrapf(err, "error getting notifications for event '%s'", v.Event().Id().Hex())
		}

		fmt.Println("event", evt.Id().Hex())

		for _, notification := range notifications {
			fmt.Println("notification", notification.UUID())
		}

	}

	return nil
}
