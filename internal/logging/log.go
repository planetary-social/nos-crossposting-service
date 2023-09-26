package logging

const (
	loggerFieldName  = "name"
	loggerFieldError = "error"
)

type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelError
	LevelDisabled
)

type Logger interface {
	New(name string) Logger
	WithError(err error) Logger
	WithField(key string, v any) Logger

	Error() Entry
	Debug() Entry
	Trace() Entry
}

type Entry interface {
	WithError(err error) Entry
	WithField(key string, v any) Entry
	Message(msg string)
}

type LoggingSystem interface {
	EnabledLevel() Level
	Error() LoggingSystemEntry
	Debug() LoggingSystemEntry
	Trace() LoggingSystemEntry
}

type LoggingSystemEntry interface {
	WithField(key string, v any) LoggingSystemEntry
	Message(msg string)
}

type SystemLogger struct {
	fields map[string]any
	logger LoggingSystem
}

func NewSystemLogger(logger LoggingSystem, name string) Logger {
	if logger.EnabledLevel() >= LevelDisabled {
		return NewDevNullLogger()
	}
	return newSystemLogger(logger, map[string]any{loggerFieldName: name})
}

func newSystemLogger(logger LoggingSystem, fields map[string]any) SystemLogger {
	newLogger := SystemLogger{
		fields: make(map[string]any),

		logger: logger,
	}

	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

func (l SystemLogger) Error() Entry {
	if l.logger.EnabledLevel() > LevelError {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Error()))
}

func (l SystemLogger) Debug() Entry {
	if l.logger.EnabledLevel() > LevelDebug {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Debug()))
}

func (l SystemLogger) Trace() Entry {
	if l.logger.EnabledLevel() > LevelTrace {
		return newDevNullLoggerEntry()
	}
	return l.withFields(newEntry(l.logger.Trace()))
}

func (l SystemLogger) New(name string) Logger {
	newLogger := newSystemLogger(l.logger, l.fields)
	v, okExists := l.fields[loggerFieldName]
	if okExists {
		if stringV, okType := v.(string); okType {
			newLogger.fields[loggerFieldName] = stringV + "." + name
			return newLogger
		}
		return newLogger
	}
	newLogger.fields[loggerFieldName] = name
	return newLogger
}

func (l SystemLogger) WithError(err error) Logger {
	newLogger := newSystemLogger(l.logger, l.fields)
	newLogger.fields[loggerFieldError] = err
	return newLogger
}

func (l SystemLogger) WithField(key string, v any) Logger {
	newLogger := newSystemLogger(l.logger, l.fields)
	newLogger.fields[key] = v
	return newLogger
}

func (l SystemLogger) withFields(entry Entry) Entry {
	for k, v := range l.fields {
		entry = entry.WithField(k, v)
	}
	return entry
}

type entry struct {
	loggingSystemEntry LoggingSystemEntry
}

func newEntry(loggingSystemEntry LoggingSystemEntry) entry {
	return entry{loggingSystemEntry: loggingSystemEntry}
}

func (e entry) WithError(err error) Entry {
	return newEntry(e.loggingSystemEntry.WithField(loggerFieldError, err))
}

func (e entry) WithField(key string, v any) Entry {
	return newEntry(e.loggingSystemEntry.WithField(key, v))
}

func (e entry) Message(msg string) {
	e.loggingSystemEntry.Message(msg)
}

type DevNullLogger struct {
}

func NewDevNullLogger() DevNullLogger {
	return DevNullLogger{}
}

func (d DevNullLogger) New(name string) Logger {
	return d
}

func (d DevNullLogger) WithError(err error) Logger {
	return d
}

func (d DevNullLogger) WithField(key string, v any) Logger {
	return d
}

func (d DevNullLogger) Error() Entry {
	return newDevNullLoggerEntry()
}

func (d DevNullLogger) Debug() Entry {
	return newDevNullLoggerEntry()
}

func (d DevNullLogger) Trace() Entry {
	return newDevNullLoggerEntry()
}

type devNullLoggerEntry struct {
}

func newDevNullLoggerEntry() devNullLoggerEntry {
	return devNullLoggerEntry{}
}

func (d devNullLoggerEntry) WithError(err error) Entry {
	return d
}

func (d devNullLoggerEntry) WithField(key string, v any) Entry {
	return d
}

func (d devNullLoggerEntry) Message(msg string) {
}
