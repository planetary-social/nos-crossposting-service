package twitter

import (
	"testing"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/fixtures"
	"github.com/stretchr/testify/require"
)

func TestLimiterResetsItself(t *testing.T) {
	l := NewLimiter()

	key := fixtures.SomeString()

	for i := 0; i < 200; i++ {
		err := l.Limit(key, 100, 10*time.Millisecond)
		if err != nil {
			require.Equal(t, 101, i)
			break
		}
	}

	<-time.After(10 * time.Millisecond)

	for i := 0; i < 200; i++ {
		err := l.Limit(key, 100, 10*time.Millisecond)
		if err != nil {
			require.Equal(t, 101, i)
			break
		}
	}
}
