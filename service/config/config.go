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
	nostrListenAddress   string
	metricsListenAddress string

	firestoreProjectID       string
	firestoreCredentialsJSON []byte

	apnsTopic               string
	apnsCertificatePath     string
	apnsCertificatePassword string

	environment Environment
	logLevel    logging.Level
}

func NewConfig(
	nostrListenAddress string,
	metricsListenAddress string,
	firestoreProjectID string,
	firestoreCredentialsJSON []byte,
	apnsTopic string,
	apnsCertificatePath string,
	apnsCertificatePassword string,
	environment Environment,
	logLevel logging.Level,
) (Config, error) {
	c := Config{
		nostrListenAddress:       nostrListenAddress,
		metricsListenAddress:     metricsListenAddress,
		firestoreProjectID:       firestoreProjectID,
		firestoreCredentialsJSON: firestoreCredentialsJSON,
		apnsTopic:                apnsTopic,
		apnsCertificatePath:      apnsCertificatePath,
		apnsCertificatePassword:  apnsCertificatePassword,
		environment:              environment,
		logLevel:                 logLevel,
	}

	c.setDefaults()
	if err := c.validate(); err != nil {
		return Config{}, errors.Wrap(err, "invalid config")
	}

	return c, nil
}

func (c *Config) NostrListenAddress() string {
	return c.nostrListenAddress
}

func (c *Config) MetricsListenAddress() string {
	return c.metricsListenAddress
}

func (c *Config) FirestoreProjectID() string {
	return c.firestoreProjectID
}

func (c *Config) FirestoreCredentialsJSON() []byte {
	return c.firestoreCredentialsJSON
}

func (c *Config) APNSTopic() string {
	return c.apnsTopic
}

func (c *Config) APNSCertificatePath() string {
	return c.apnsCertificatePath
}

func (c *Config) APNSCertificatePassword() string {
	return c.apnsCertificatePassword
}

func (c *Config) Environment() Environment {
	return c.environment
}

func (c *Config) LogLevel() logging.Level {
	return c.logLevel
}

func (c *Config) setDefaults() {
	if c.nostrListenAddress == "" {
		c.nostrListenAddress = ":8008"
	}

	if c.metricsListenAddress == "" {
		c.metricsListenAddress = ":8009"
	}
}

func (c *Config) validate() error {
	if c.nostrListenAddress == "" {
		return errors.New("missing nostr listen address")
	}

	if c.metricsListenAddress == "" {
		return errors.New("missing metrics listen address")
	}

	if c.firestoreProjectID == "" {
		return errors.New("missing firestore project id")
	}

	if c.apnsTopic == "" {
		return errors.New("missing APNs topic")
	}

	if c.apnsCertificatePath == "" {
		return errors.New("missing APNs certificate path")
	}

	switch c.environment {
	case EnvironmentProduction:
	case EnvironmentDevelopment:
	default:
		return fmt.Errorf("unknown environment '%+v'", c.environment)
	}

	return nil
}
