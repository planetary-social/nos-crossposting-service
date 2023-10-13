//go:build test_integration

package integration_tests

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/cmd/crossposting-service/di"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	durationTimeout = 1 * time.Second
	durationTick    = 100 * time.Millisecond
)

func TestFlow(t *testing.T) {
	ctx := fixtures.Context(t)
	config, service := createService(ctx, t)

	publicKey, privateKeyHex := fixtures.SomeKeyPair()
	token := fixtures.SomeAPNSToken()

	env := testEnvironment{
		config:            config,
		service:           service,
		registerPublicKey: publicKey,
		registerSecretKey: privateKeyHex,
		token:             token,
	}

	testAddRegistration(t, ctx, env)
	testIngestEventAndSendOutNotifications(t, ctx, env)
}

type testEnvironment struct {
	config  config.Config
	service di.IntegrationService

	registerPublicKey domain.PublicKey
	registerSecretKey string
	token             domain.APNSToken
}

func testAddRegistration(t *testing.T, ctx context.Context, env testEnvironment) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn := createClient(ctx, t, env.config)

	relayAddress := fixtures.SomeRelayAddress()

	event := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      12345,
		Tags:      nostr.Tags{},
		Content: fmt.Sprintf(`
{
  "publicKey": "%s",
  "relays": [
	{
	  "address": "%s"
	}
  ],
  "apnsToken": "%s"
}
`,
			env.registerPublicKey.Hex(),
			relayAddress.String(),
			env.token.Hex(),
		),
	}

	err := event.Sign(env.registerSecretKey)
	require.NoError(t, err)

	envelope := nostr.EventEnvelope{
		SubscriptionID: nil,
		Event:          event,
	}

	j, err := envelope.MarshalJSON()
	require.NoError(t, err)

	err = conn.WriteMessage(websocket.TextMessage, j)
	require.NoError(t, err)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		relays, err := env.service.Service.App().Queries.GetRelays.Handle(ctx)
		assert.NoError(c, err)
		assert.Contains(c, relays, relayAddress)
	}, durationTimeout, durationTick)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		publicKeys, err := env.service.Service.App().Queries.GetPublicKeys.Handle(ctx, relayAddress)
		assert.NoError(c, err)
		assert.Contains(c, publicKeys, env.registerPublicKey)
	}, durationTimeout, durationTick)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		tokens, err := env.service.Service.App().Queries.GetTokens.Handle(ctx, env.registerPublicKey)
		assert.NoError(c, err)
		assert.Contains(c, tokens, env.token)
	}, durationTimeout, durationTick)
}

func testIngestEventAndSendOutNotifications(t *testing.T, ctx context.Context, env testEnvironment) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	_, otherPersonSecretKey := fixtures.SomeKeyPair()

	libevent := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      domain.EventKindNote.Int(),
		Tags: nostr.Tags{
			{"p", env.registerPublicKey.Hex()},
		},
		Content: "some content",
	}

	err := libevent.Sign(otherPersonSecretKey)
	require.NoError(t, err)

	event, err := domain.NewEvent(libevent)
	require.NoError(t, err)

	cmd := app.NewProcessReceivedEvent(fixtures.SomeRelayAddress(), event)
	err = env.service.Service.App().Commands.SaveReceivedEvent.Handle(ctx, cmd)
	require.NoError(t, err)

	// event is eventually available through websockets
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		someSubscriptionId := "some-subscription-id"
		client := createClient(ctx, t, env.config)

		since := nostr.Timestamp(time.Now().Add(-1 * time.Hour).Unix())

		filters := nostr.Filters{
			{
				Kinds: []int{domain.EventKindNote.Int(), domain.EventKindReaction.Int()},
				Tags: nostr.TagMap{
					"p": []string{env.registerPublicKey.Hex()},
				},
				Since: &since,
			},
		}

		envelope := nostr.ReqEnvelope{
			SubscriptionID: someSubscriptionId,
			Filters:        filters,
		}

		err = client.WriteJSON(envelope)
		require.NoError(t, err)

		var events []string

	loop:
		for {
			_, msg, err := client.ReadMessage()
			require.NoError(t, err)

			switch v := nostr.ParseMessage(msg).(type) {
			case *nostr.EventEnvelope:
				fmt.Println(string(msg), v)
				events = append(events, string(msg))
			case *nostr.EOSEEnvelope:
				break loop
			case nil:
				t.Fatal("nil")
			default:
				t.Fatal("default")
			}

		}

		eventEnvelope := nostr.EventEnvelope{
			SubscriptionID: &someSubscriptionId,
			Event:          libevent,
		}

		eventEnvelopeJSON, err := eventEnvelope.MarshalJSON()
		require.NoError(t, err)

		assert.Equal(t, len(events), 1)
		assert.Equal(t,
			strings.TrimSpace(string(eventEnvelopeJSON)),
			strings.TrimSpace(events[0]),
		)
	}, durationTimeout, durationTick)

	// event triggered notifications
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.Greater(t, len(env.service.MockAPNS.SentNotifications()), 0)

		for _, notification := range env.service.MockAPNS.SentNotifications() {
			assert.Equal(t, env.token, notification.APNSToken())
			assert.Equal(t, `{"aps":{"content-available":1}}`, string(notification.Payload()))

		}
	}, durationTimeout, durationTick)

	// notifications were persisted
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		notifications, err := env.service.Service.App().Queries.GetNotifications.Handle(ctx, event.Id())
		require.NoError(t, err)

		assert.Greater(t, len(notifications), 0)
	}, durationTimeout, durationTick)
}

func createClient(ctx context.Context, tb testing.TB, conf config.Config) *websocket.Conn {
	addr := conf.NostrListenAddress()
	if strings.HasPrefix(addr, ":") {
		addr = "localhost" + addr
	}
	addr = fmt.Sprintf("ws://%s", addr)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, addr, nil)
	require.NoError(tb, err)
	return conn
}

func createService(ctx context.Context, tb testing.TB) (config.Config, di.IntegrationService) {
	config, err := config.NewConfig(
		fmt.Sprintf(":%d", 8000+rand.Int()%1000),
		fmt.Sprintf(":%d", 8000+rand.Int()%1000),
		"test-project-id",
		nil,
		"someAPNSTopic",
		"someAPNSCertPath",
		"someAPNSCertPassword",
		config.EnvironmentDevelopment,
		logging.LevelTrace,
	)
	require.NoError(tb, err)

	service, cleanup, err := di.BuildIntegrationService(ctx, config)
	require.NoError(tb, err)
	tb.Cleanup(cleanup)

	terminatedCh := make(chan error)
	tb.Cleanup(func() {
		if err := <-terminatedCh; err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			tb.Fatalf("error shutting down the service: %s", err)
		}
	})

	runCtx, cancelRunCtx := context.WithCancel(ctx)
	tb.Cleanup(cancelRunCtx)
	go func() {
		terminatedCh <- service.Service.Run(runCtx)
	}()

	return config, service
}
