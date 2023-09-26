package logging

import "github.com/ThreeDotsLabs/watermill"

type WatermillAdapter struct {
	logger Logger
}

func NewWatermillAdapter(logger Logger) *WatermillAdapter {
	return &WatermillAdapter{logger: logger}
}

func (a WatermillAdapter) Error(msg string, err error, fields watermill.LogFields) {
	a.withFields(a.logger.Error(), fields).WithError(err).Message(msg)
}

func (a WatermillAdapter) Info(msg string, fields watermill.LogFields) {
	a.withFields(a.logger.Debug(), fields).Message(msg)
}

func (a WatermillAdapter) Debug(msg string, fields watermill.LogFields) {
	a.withFields(a.logger.Debug(), fields).Message(msg)
}

func (a WatermillAdapter) Trace(msg string, fields watermill.LogFields) {
	a.withFields(a.logger.Trace(), fields).Message(msg)
}

func (a WatermillAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return NewWatermillAdapter(a.loggerWithFields(a.logger, fields))
}

func (a WatermillAdapter) withFields(e Entry, fields watermill.LogFields) Entry {
	for name, value := range fields {
		e = e.WithField(name, value)
	}
	return e
}

func (a WatermillAdapter) loggerWithFields(e Logger, fields watermill.LogFields) Logger {
	for name, value := range fields {
		e = e.WithField(name, value)
	}
	return e
}
