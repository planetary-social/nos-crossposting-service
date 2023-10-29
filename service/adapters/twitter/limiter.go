package twitter

import (
	"time"

	"github.com/boreq/errors"
)

type Limiter struct {
	m map[string][]time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{
		m: make(map[string][]time.Time),
	}
}

func (l *Limiter) Limit(key string, number int, window time.Duration) error {
	for i := 0; i < len(l.m[key]); i++ {
		if time.Since(l.m[key][i]) > window {
			l.m[key] = append(l.m[key][:i], l.m[key][i+1:]...)
		}
	}

	if len(l.m[key]) > number {
		return errors.New("exceeded the limit")
	}

	l.m[key] = append(l.m[key], time.Now())
	return nil
}
