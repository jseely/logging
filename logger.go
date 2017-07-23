package logging

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jseely/logging/common"
	"github.com/jseely/utils"
)

type logger struct {
	scope    string
	minLevel common.Level
	sinks    []common.Sink
}

type Logger interface {
	WithApplicationScope(scope string, minLevel common.Level) Logger
	Close() error
	Verbose(message string, values ...interface{})
	Debug(message string, values ...interface{})
	Information(message string, values ...interface{})
	Warning(message string, values ...interface{})
	Error(message string, values ...interface{})
	Fatal(message string, values ...interface{})
}

func New(minLevel common.Level, sinks ...common.Sink) Logger {
	return &logger{
		minLevel: minLevel,
		sinks:    sinks,
	}
}

func NewWithApplicationScope(scope string, minLevel common.Level, sinks ...common.Sink) Logger {
	return &logger{
		scope:    scope,
		minLevel: minLevel,
		sinks:    sinks,
	}
}

func (l *logger) WithApplicationScope(scope string, minLevel common.Level) Logger {
	return &logger{
		scope:    l.scope + "." + scope,
		minLevel: minLevel,
		sinks:    l.sinks,
	}
}

func (l *logger) Close() error {
	errString := ""
	for _, sink := range l.sinks {
		err := sink.Close()
		if err != nil {
			if errString == "" {
				errString = err.Error()
			} else {
				errString += " | " + err.Error()
			}
		}
	}
	if errString == "" {
		return nil
	} else {
		return fmt.Errorf("Failed to close one or more sinks. %s", errString)
	}
}

func (l *logger) Verbose(message string, values ...interface{}) {
	l.dispatch(common.VERBOSE, message, values)
}

func (l *logger) Debug(message string, values ...interface{}) {
	l.dispatch(common.DEBUG, message, values)
}

func (l *logger) Information(message string, values ...interface{}) {
	l.dispatch(common.INFORMATION, message, values)
}

func (l *logger) Warning(message string, values ...interface{}) {
	l.dispatch(common.WARNING, message, values)
}

func (l *logger) Error(message string, values ...interface{}) {
	l.dispatch(common.ERROR, message, values)
}

func (l *logger) Fatal(message string, values ...interface{}) {
	l.dispatch(common.FATAL, message, values)
}

func (l *logger) dispatch(level common.Level, messageTemplate string, values []interface{}) {
	if level < l.minLevel {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			v, _ := json.Marshal(values)
			l.Error("Panicking in {methodName}, writing {level} message with template \"{messageTemplate}\" and values {values}. {panicMessage} {panicIdentity}", "logging.dispatch", level.String(), messageTemplate, string(v), r, utils.IdentifyPanic())
		}
	}()
	if l == nil {
		return
	}
	fields := createFieldsMap(messageTemplate, values)
	for _, sink := range l.sinks {
		sink.Write(l.scope, level, messageTemplate, fields)
	}
}

var parser = regexp.MustCompile(`{[a-zA-Z0-9]+}`)

func createFieldsMap(message string, values []interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	i := 0
	for _, k := range parser.FindAllString(message, -1) {
		k = strings.TrimPrefix(k, "{")
		k = strings.TrimSuffix(k, "}")
		if _, ok := fields[k]; !ok {
			fields[k] = values[i]
			i = i + 1
		}
	}
	return fields
}
