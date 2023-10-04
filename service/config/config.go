package config

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
)

type Environment struct {
	s string
}

var (
	EnvironmentProduction  = Environment{"production"}
	EnvironmentDevelopment = Environment{"development"}
)

type Config struct {
	listenAddress        string
	metricsListenAddress string

	environment Environment
	logLevel    logging.Level

	twitterKey       string
	twitterKeySecret string
}

func NewConfig(
	listenAddress string,
	metricsListenAddress string,
	environment Environment,
	logLevel logging.Level,
	twitterKey string,
	twitterKeySecret string,
) (Config, error) {
	c := Config{
		listenAddress:        listenAddress,
		metricsListenAddress: metricsListenAddress,
		environment:          environment,
		logLevel:             logLevel,
		twitterKey:           twitterKey,
		twitterKeySecret:     twitterKeySecret,
	}

	c.setDefaults()
	if err := c.validate(); err != nil {
		return Config{}, errors.Wrap(err, "invalid config")
	}

	return c, nil
}

func (c *Config) ListenAddress() string {
	return c.listenAddress
}

func (c *Config) MetricsListenAddress() string {
	return c.metricsListenAddress
}

func (c *Config) Environment() Environment {
	return c.environment
}

func (c *Config) LogLevel() logging.Level {
	return c.logLevel
}

func (c *Config) TwitterKey() string {
	return c.twitterKey
}

func (c *Config) TwitterKeySecret() string {
	return c.twitterKeySecret
}

func (c *Config) setDefaults() {
	if c.listenAddress == "" {
		c.listenAddress = ":8008"
	}

	if c.metricsListenAddress == "" {
		c.metricsListenAddress = ":8009"
	}
}

func (c *Config) validate() error {
	if c.listenAddress == "" {
		return errors.New("missing listen address")
	}

	if c.metricsListenAddress == "" {
		return errors.New("missing metrics listen address")
	}

	switch c.environment {
	case EnvironmentProduction:
	case EnvironmentDevelopment:
	default:
		return fmt.Errorf("unknown environment '%+v'", c.environment)
	}

	if c.twitterKey == "" {
		return errors.New("missing twitter key")
	}

	if c.twitterKeySecret == "" {
		return errors.New("missing twitter key secret")
	}

	return nil
}
