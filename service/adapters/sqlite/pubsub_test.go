package sqlite_test

import (
	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
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

func TestPubSub_AckedMessagesAreNotRetried(t *testing.T) {
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
			err := msg.Ack()
			require.NoError(t, err)
		}
	}()

	<-time.After(10 * time.Second)
	msgsLock.Lock()
	require.Len(t, msgs, 1)
	msgsLock.Unlock()
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
