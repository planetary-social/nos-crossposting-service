package app_test

import (
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/cmd/crossposting-service/di"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/stretchr/testify/require"
)

func TestSendTweetHandler_CorrectlyDropsOldEvents(t *testing.T) {
	testCases := []struct {
		Name string

		CurrentTime time.Time
		Event       *domain.Event

		ShouldPostTweet bool
	}{
		{
			Name: "old_events_are_dropped_after_a_week_since_code_change_passes",

			CurrentTime: date(2023, time.November, 25),
			Event:       nil,

			ShouldPostTweet: false,
		},
		{
			Name: "old_events_are_not_dropped_before_a_week_since_code_change_passes",

			CurrentTime: date(2023, time.November, 23),
			Event:       nil,

			ShouldPostTweet: true,
		},
		{
			Name: "new_events_are_dropped_after_a_week_since_they_were_created",

			CurrentTime: date(2023, time.November, 28),
			Event:       internal.Pointer(fixtures.SomeEventWithCreatedAt(date(2023, time.November, 20))),

			ShouldPostTweet: false,
		},
		{
			Name: "new_events_are_not_dropped_before_a_week_since_they_were_created",

			CurrentTime: date(2023, time.November, 27),
			Event:       internal.Pointer(fixtures.SomeEventWithCreatedAt(date(2023, time.November, 20))),

			ShouldPostTweet: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ts, err := di.BuildTestApplication(t)
			require.NoError(t, err)

			ctx := fixtures.TestContext(t)

			accountId := fixtures.SomeAccountID()
			tweet := domain.NewTweet(fixtures.SomeString())
			userTokens := accounts.NewTwitterUserTokens(
				accountId,
				fixtures.SomeTwitterUserAccessToken(),
				fixtures.SomeTwitterUserAccessSecret(),
			)
			ts.UserTokensRepository.MockUserTokens(userTokens)
			ts.CurrentTimeProvider.SetCurrentTime(testCase.CurrentTime)

			cmd := app.NewSendTweet(accountId, tweet, testCase.Event)

			err = ts.SendTweetHandler.Handle(ctx, cmd)
			require.NoError(t, err)

			if testCase.ShouldPostTweet {
				require.Len(t, ts.Twitter.PostTweetCalls, 1)
			} else {
				require.Len(t, ts.Twitter.PostTweetCalls, 0)
			}
		})
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
