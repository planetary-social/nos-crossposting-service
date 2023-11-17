package mocks

import "time"

type CurrentTimeProvider struct {
	CurrentTime time.Time
}

func NewCurrentTimeProvider() *CurrentTimeProvider {
	return &CurrentTimeProvider{}
}

func (c *CurrentTimeProvider) GetCurrentTime() time.Time {
	return c.CurrentTime
}

func (c *CurrentTimeProvider) SetCurrentTime(currentTime time.Time) {
	c.CurrentTime = currentTime
}
