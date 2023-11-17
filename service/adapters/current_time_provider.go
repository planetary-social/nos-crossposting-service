package adapters

import "time"

type CurrentTimeProvider struct {
}

func NewCurrentTimeProvider() *CurrentTimeProvider {
	return &CurrentTimeProvider{}
}

func (c CurrentTimeProvider) GetCurrentTime() time.Time {
	return time.Now()
}
