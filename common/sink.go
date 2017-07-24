package common

type Sink interface {
	Write(scope string, level Level, messageTemplate string, fields map[string]interface{})
	Close() error
}
