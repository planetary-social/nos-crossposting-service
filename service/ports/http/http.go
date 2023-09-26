package http

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/boreq/errors"
	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/config"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
)

type Server struct {
	config config.Config
	app    app.Application
	logger logging.Logger
}

func NewServer(
	config config.Config,
	app app.Application,
	logger logging.Logger,
) Server {
	return Server{
		config: config,
		app:    app,
		logger: logger.New("server"),
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := s.createMux(ctx)

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", s.config.NostrListenAddress())
	if err != nil {
		return errors.Wrap(err, "error listening")
	}

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			fmt.Println("error closing listener:", err)
		}
	}()

	return http.Serve(listener, mux)
}

func (s *Server) createMux(ctx context.Context) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(ctx, w, r)
	})
	return mux
}

func (s *Server) serveWs(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		s.logger.Error().WithError(err).Message("error upgrading the connection")
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Error().WithError(err).Message("error closing the connection")
		}
	}()

	if err := s.handleConnection(ctx, conn); err != nil {
		closeErr := &websocket.CloseError{}
		if !errors.As(err, &closeErr) || closeErr.Code != websocket.CloseNormalClosure {
			s.logger.Error().WithError(err).Message("error handling the connection")
		}
	}
}

func (s *Server) handleConnection(ctx context.Context, conn *websocket.Conn) error {
	s.logger.Debug().Message("accepted websocket connection")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	subscriptions := make(map[string]context.CancelFunc)

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			return errors.Wrap(err, "error reading the websocket message")
		}

		message := nostr.ParseMessage(messageBytes)
		if message == nil {
			return errors.New("failed to parse the message")
		}

		switch v := message.(type) {
		case *nostr.EventEnvelope:
			event, err := domain.NewEvent(v.Event)
			if err != nil {
				return errors.Wrap(err, "error creating an event")
			}

			registration, err := domain.NewRegistrationFromEvent(event)
			if err != nil {
				return errors.Wrap(err, "error creating a registration")
			}

			cmd := app.NewSaveRegistration(
				registration,
			)

			if err := s.app.Commands.SaveRegistration.Handle(ctx, cmd); err != nil {
				return errors.Wrap(err, "error handling the registration command")
			}
		case *nostr.ReqEnvelope:
			filters, err := domain.NewFilters(v.Filters)
			if err != nil {
				return errors.Wrap(err, "error creating filters")
			}

			s.closeSubscription(subscriptions, v.SubscriptionID)

			subCtx, subCancel := context.WithCancel(ctx)
			go s.sendEvents(subCtx, conn, filters, v.SubscriptionID)
			subscriptions[v.SubscriptionID] = subCancel
		case *nostr.CloseEnvelope:
			s.closeSubscription(subscriptions, string(*v))
		default:
			s.logger.Error().WithField("message", message).Message("received an unknown message")
		}
	}
}

func (s *Server) sendEvents(ctx context.Context, conn *websocket.Conn, filters domain.Filters, subscriptionName string) {
	if err := s.sendEventsErr(ctx, conn, filters, subscriptionName); err != nil {
		if !errors.Is(err, context.Canceled) {
			s.logger.Error().WithError(err).Message("get events returned an error")
			return
		}
	}
}

func (s *Server) sendEventsErr(ctx context.Context, conn *websocket.Conn, filters domain.Filters, subscriptionName string) error {
	for event := range s.app.Queries.GetEvents.Handle(ctx, filters) {
		if err := event.Err(); err != nil {
			return errors.Wrap(err, "received an error")
		}

		if event.EOSE() {
			envelope := nostr.EOSEEnvelope(subscriptionName)

			if err := conn.WriteJSON(envelope); err != nil {
				return errors.Wrap(err, "error writing EOSE")
			}

			continue
		}

		envelope := nostr.EventEnvelope{
			SubscriptionID: &subscriptionName,
			Event:          event.Event().Libevent(),
		}

		if err := conn.WriteJSON(envelope); err != nil {
			return errors.Wrap(err, "error writing an event")
		}
	}

	return nil
}

func (s *Server) closeSubscription(subscriptions map[string]context.CancelFunc, subscriptionName string) {
	if cancel, ok := subscriptions[subscriptionName]; ok {
		cancel()
		delete(subscriptions, subscriptionName)
	}
}
