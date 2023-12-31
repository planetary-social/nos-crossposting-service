package fixtures

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/planetary-social/nos-crossposting-service/internal"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
	"github.com/planetary-social/nos-crossposting-service/service/domain/sessions"
)

func somePrivateKeyHex() string {
	return nostr.GeneratePrivateKey()
}

func SomeKeyPair() (publicKey domain.PublicKey, secretKeyHex string) {
	hex := somePrivateKeyHex()

	p, err := nostr.GetPublicKey(hex)
	if err != nil {
		panic(err)
	}
	v, err := domain.NewPublicKeyFromHex(p)
	if err != nil {
		panic(err)
	}
	return v, hex
}

func SomePublicKey() domain.PublicKey {
	hex := somePrivateKeyHex()

	p, err := nostr.GetPublicKey(hex)
	if err != nil {
		panic(err)
	}
	v, err := domain.NewPublicKeyFromHex(p)
	if err != nil {
		panic(err)
	}
	return v
}

func SomeRelayAddress() domain.RelayAddress {
	protocol := internal.RandomElement([]string{"ws", "wss"})
	address := fmt.Sprintf("%s://%s", protocol, SomeString())

	v, err := domain.NewRelayAddress(address)
	if err != nil {
		panic(err)
	}
	return v
}

func SomeString() string {
	return randSeq(10)
}

func SomeEventID() domain.EventId {
	return domain.MustNewEventId(SomeHexBytesOfLen(32))
}

func SomeAccountID() accounts.AccountID {
	return accounts.MustNewAccountID(SomeHexBytesOfLen(10))
}

func SomeSessionID() sessions.SessionID {
	return sessions.MustNewSessionID(SomeHexBytesOfLen(10))
}

func SomeTwitterID() accounts.TwitterID {
	return accounts.NewTwitterID(rand.Int63())
}

func SomeHexBytesOfLen(l int) string {
	b := make([]byte, l)
	n, err := cryptorand.Read(b)
	if n != len(b) {
		panic("short read")
	}
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func SomeBytesOfLen(l int) []byte {
	b := make([]byte, l)
	n, err := cryptorand.Read(b)
	if n != len(b) {
		panic("short read")
	}
	if err != nil {
		panic(err)
	}
	return b
}

func SomeFile(t testing.TB) string {
	file, err := os.CreateTemp("", "nos-crossposting-test")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return file.Name()
}

func TestContext(t testing.TB) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

func SomeError() error {
	return fmt.Errorf("some error: %d", rand.Int())
}

func SomeTwitterUserAccessToken() accounts.TwitterUserAccessToken {
	v, err := accounts.NewTwitterUserAccessToken(SomeString())
	if err != nil {
		panic(err)
	}
	return v
}

func SomeTwitterUserAccessSecret() accounts.TwitterUserAccessSecret {
	v, err := accounts.NewTwitterUserAccessSecret(SomeString())
	if err != nil {
		panic(err)
	}
	return v
}

func SomeEventWithCreatedAt(createdAt time.Time) domain.Event {
	_, sk := SomeKeyPair()

	libevent := nostr.Event{
		CreatedAt: nostr.Timestamp(createdAt.Unix()),
		Kind:      domain.EventKindNote.Int(),
		Content:   SomeString(),
	}
	err := libevent.Sign(sk)
	if err != nil {
		panic(err)
	}

	event, err := domain.NewEvent(libevent)
	if err != nil {
		panic(err)
	}

	return event
}

func SomeEvent() domain.Event {
	return SomeEventWithCreatedAt(time.Now())
}

func TestLogger(tb testing.TB) logging.Logger {
	return logging.NewSystemLogger(logging.NewTestingLoggingSystem(tb), "test")
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
