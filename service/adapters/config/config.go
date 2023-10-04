package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/config"
)

const (
	envPrefix = "NOTIFICATIONS"

	envNostrListenAddress   = "NOSTR_LISTEN_ADDRESS"
	envMetricsListenAddress = "METRICS_LISTEN_ADDRESS"
	envEnvironment          = "ENVIRONMENT"
	envLogLevel             = "LOG_LEVEL"
	envTwitterKey           = "TWITTER_KEY"
	envTwitterKeySecret     = "TWITTER_KEY_SECRET"
)

type EnvironmentConfigLoader struct {
}

func NewEnvironmentConfigLoader() *EnvironmentConfigLoader {
	return &EnvironmentConfigLoader{}
}

func (c *EnvironmentConfigLoader) Load() (config.Config, error) {
	environment, err := c.loadEnvironment()
	if err != nil {
		return config.Config{}, errors.Wrap(err, "error loading the environment setting")
	}

	logLevel, err := c.loadLogLevel()
	if err != nil {
		return config.Config{}, errors.Wrap(err, "error loading the log level")
	}

	return config.NewConfig(

		c.getenv(envNostrListenAddress),
		c.getenv(envMetricsListenAddress),
		environment,
		logLevel,
		c.getenv(envTwitterKey),
		c.getenv(envTwitterKeySecret),
	)
}

func (c *EnvironmentConfigLoader) loadEnvironment() (config.Environment, error) {
	v := strings.ToUpper(c.getenv(envEnvironment))
	switch v {
	case "PRODUCTION":
		return config.EnvironmentProduction, nil
	case "DEVELOPMENT":
		return config.EnvironmentDevelopment, nil
	case "":
		return config.EnvironmentProduction, nil
	default:
		return config.Environment{}, fmt.Errorf("invalid environment requested '%s'", v)
	}
}

func (c *EnvironmentConfigLoader) loadLogLevel() (logging.Level, error) {
	v := strings.ToUpper(c.getenv(envLogLevel))
	switch v {
	case "TRACE":
		return logging.LevelTrace, nil
	case "DEBUG":
		return logging.LevelDebug, nil
	case "ERROR":
		return logging.LevelError, nil
	case "DISABLED":
		return logging.LevelDisabled, nil
	case "":
		return logging.LevelDebug, nil
	default:
		return 0, fmt.Errorf("invalid log level requested '%s'", v)
	}
}

func (c *EnvironmentConfigLoader) getenv(key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", envPrefix, key))
}
