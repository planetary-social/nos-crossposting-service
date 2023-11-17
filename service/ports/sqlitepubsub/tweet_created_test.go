package sqlitepubsub

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTweetCreatedEventSubscriber_CanHandleOldAndNewEvents(t *testing.T) {
	event := fixtures.SomeEvent()

	testCases := []struct {
		Name            string
		Payload         string
		ExpectedCommand app.SendTweet
	}{
		{
			Name:    "old",
			Payload: `{"accountID": "someAccountID", "tweet": {"text": "someTweetText"}}`,
			ExpectedCommand: app.NewSendTweet(
				accounts.MustNewAccountID("someAccountID"),
				domain.NewTweet("someTweetText"),
				nil,
			),
		},
		{
			Name: "new",
			Payload: fmt.Sprintf(
				`{"accountID": "someAccountID", "tweet": {"text": "someTweetText"}, "event": "%s", "createdAt": "%s"}`,
				base64.StdEncoding.EncodeToString(event.Raw()),
				time.Now().Format(time.RFC3339),
			),
			ExpectedCommand: app.NewSendTweet(
				accounts.MustNewAccountID("someAccountID"),
				domain.NewTweet("someTweetText"),
				&event,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			h := newSendTweetHandlerMock()
			s := newSqliteSubscriberMock()
			logger := fixtures.TestLogger(t)
			subscriber := NewTweetCreatedEventSubscriber(h, s, logger)

			go func() {
				_ = subscriber.Run(ctx)
			}()

			message, err := sqlite.NewMessage(fixtures.SomeString(), []byte(testCase.Payload))
			require.NoError(t, err)

			receivedMessage := sqlite.NewReceivedMessage(message)

			err = s.PublishTweetCreated(ctx, receivedMessage)
			require.NoError(t, err)

			require.EventuallyWithT(t, func(t *assert.CollectT) {
				calls := h.Calls()
				if assert.Len(t, calls, 1) {
					call := calls[0]
					assert.Equal(t, call.AccountID(), testCase.ExpectedCommand.AccountID())
					assert.Equal(t, call.Tweet(), testCase.ExpectedCommand.Tweet())
					if testCase.ExpectedCommand.Event() == nil {
						require.Nil(t, call.Event())
					} else {
						assert.Equal(t, call.Event().Raw(), testCase.ExpectedCommand.Event().Raw())
					}
				}
			}, 1*time.Second, 100*time.Millisecond)
		})
	}
}

type sendTweetHandlerMock struct {
	calls     []app.SendTweet
	callsLock sync.Mutex
}

func newSendTweetHandlerMock() *sendTweetHandlerMock {
	return &sendTweetHandlerMock{}
}

func (s *sendTweetHandlerMock) Handle(ctx context.Context, cmd app.SendTweet) (err error) {
	s.callsLock.Lock()
	defer s.callsLock.Unlock()
	s.calls = append(s.calls, cmd)
	return nil
}

func (s *sendTweetHandlerMock) Calls() []app.SendTweet {
	s.callsLock.Lock()
	defer s.callsLock.Unlock()
	return internal.CopySlice(s.calls)
}

type sqliteSubscriberMock struct {
	ch chan *sqlite.ReceivedMessage
}

func newSqliteSubscriberMock() *sqliteSubscriberMock {
	return &sqliteSubscriberMock{
		ch: make(chan *sqlite.ReceivedMessage),
	}
}

func (s *sqliteSubscriberMock) SubscribeToTweetCreated(ctx context.Context) <-chan *sqlite.ReceivedMessage {
	return s.ch
}

func (s *sqliteSubscriberMock) PublishTweetCreated(ctx context.Context, message *sqlite.ReceivedMessage) error {
	select {
	case s.ch <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
