package main

import (
	"context"
	"fmt"
	"os"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/cmd/crossposting-service/di"
	configadapters "github.com/planetary-social/nos-crossposting-service/service/adapters/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	conf, err := configadapters.NewEnvironmentConfigLoader().Load()
	if err != nil {
		return errors.Wrap(err, "error creating a config")
	}

	service, cleanup, err := di.BuildService(ctx, conf)
	if err != nil {
		return errors.Wrap(err, "error building a service")
	}
	defer cleanup()

	if err := service.ExecuteMigrations(ctx); err != nil {
		return errors.Wrap(err, "error executing migrations")
	}

	return service.Run(ctx)
}
