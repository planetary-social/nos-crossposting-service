package sqlite_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub_PublishDoesNotReturnErrors(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	msg, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	err = adapters.PubSub.Publish(fixtures.SomeString(), msg)
	require.NoError(t, err)
}

func TestPubSub_PublishingMessagesWithIdenticalUUIDsReturnsAnError(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	msg, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	err = adapters.PubSub.Publish(fixtures.SomeString(), msg)
	require.NoError(t, err)

	err = adapters.PubSub.Publish(fixtures.SomeString(), msg)
	require.EqualError(t, err, "UNIQUE constraint failed: pubsub.uuid")
}

func TestPubSub_NackedMessagesAreRetried(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	msg, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	topic := fixtures.SomeString()

	err = adapters.PubSub.Publish(topic, msg)
	require.NoError(t, err)

	var msgs []*sqlite.ReceivedMessage
	var msgsLock sync.Mutex

	go func() {
		for msg := range adapters.PubSub.Subscribe(ctx, topic) {
			msgsLock.Lock()
			msgs = append(msgs, msg)
			msgsLock.Unlock()
			err := msg.Nack()
			require.NoError(t, err)
		}
	}()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		msgsLock.Lock()
		assert.GreaterOrEqual(collect, len(msgs), 2)
		msgsLock.Unlock()
	}, 10*time.Second, 100*time.Microsecond)
}

func TestPubSub_MessageContainCorrectPayloadAndAckedMessagesAreNotRetried(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		Payload []byte
	}{
		{
			Name:    "nil",
			Payload: nil,
		},
		{
			Name:    "not_nil",
			Payload: fixtures.SomeBytesOfLen(10),
		},
	}

	for i := range testCases {
		testCase := testCases[i]

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			ctx := fixtures.TestContext(t)
			adapters := NewTestAdapters(ctx, t)

			msg, err := sqlite.NewMessage(fixtures.SomeString(), testCase.Payload)
			require.NoError(t, err)

			topic := fixtures.SomeString()

			err = adapters.PubSub.Publish(topic, msg)
			require.NoError(t, err)

			var msgs []*sqlite.ReceivedMessage
			var msgsLock sync.Mutex

			go func() {
				for msg := range adapters.PubSub.Subscribe(ctx, topic) {
					msgsLock.Lock()
					msgs = append(msgs, msg)
					msgsLock.Unlock()
					err := msg.Ack()
					require.NoError(t, err)
				}
			}()

			<-time.After(10 * time.Second)
			msgsLock.Lock()
			require.Len(t, msgs, 1)
			require.Equal(t, msg.UUID(), msgs[0].UUID())
			require.Equal(t, msg.Payload(), msgs[0].Payload())
			msgsLock.Unlock()
		})
	}
}

func TestPubSub_NotAckedOrNackedMessagesBlock(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	msg, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	topic := fixtures.SomeString()

	err = adapters.PubSub.Publish(topic, msg)
	require.NoError(t, err)

	var msgs []*sqlite.ReceivedMessage
	var msgsLock sync.Mutex

	go func() {
		for msg := range adapters.PubSub.Subscribe(ctx, topic) {
			msgsLock.Lock()
			msgs = append(msgs, msg)
			msgsLock.Unlock()
		}
	}()

	<-time.After(10 * time.Second)
	msgsLock.Lock()
	require.Len(t, msgs, 1)
	msgsLock.Unlock()
}

func TestPubSub_QueueLengthReportsNumberOfElementsInQueue(t *testing.T) {
	t.Parallel()

	ctx := fixtures.TestContext(t)
	adapters := NewTestAdapters(ctx, t)

	msg1, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	msg2, err := sqlite.NewMessage(fixtures.SomeString(), nil)
	require.NoError(t, err)

	topic := fixtures.SomeString()

	err = adapters.PubSub.Publish(topic, msg1)
	require.NoError(t, err)

	n, err := adapters.PubSub.QueueLength(topic)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	err = adapters.PubSub.Publish(topic, msg2)
	require.NoError(t, err)

	n, err = adapters.PubSub.QueueLength(topic)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for msg := range adapters.PubSub.Subscribe(ctx, topic) {
			err := msg.Ack()
			require.NoError(t, err)
			return
		}
	}()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		n, err = adapters.PubSub.QueueLength(topic)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	}, 10*time.Second, 100*time.Millisecond)
}

func TestDefaultBackoffManager_GetMessageErrorBackoffStatisticallyFallsWithinCertainEpsilon(t *testing.T) {
	const numSamples = 1000

	m := sqlite.NewDefaultBackoffManager()
	for i := 1; i < 10; i++ {
		var sum float64
		var avg float64

		for samples := 0; samples < numSamples; samples++ {
			backoff := m.GetMessageErrorBackoff(i)
			if samples > numSamples/2 {
				require.InEpsilonf(t, avg, backoff, 0.15, "failed for i=%d and sample=%d", i, samples)
			}

			sum += float64(backoff)
			avg = sum / float64(samples)
		}

		t.Log(i, time.Duration(avg))
	}
}
