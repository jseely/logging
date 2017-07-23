package common

type Sink interface {
	Level(minLevel Level)
	Write(scope string, level Level, messageTemplate string, fields map[string]interface{})
	Close() error
}
