package sinks

import (
	"fmt"

	"github.com/jseely/logging/common"
	"golang.org/x/sys/windows/svc/eventlog"
)

type eventLogSink struct {
	log *eventlog.Log
}

func NewEventLogSink(source string) (common.Sink, error) {
	log, err := eventlog.Open(source)
	if err != nil {
		err = eventlog.InstallAsEventCreate(source, eventlog.Error|eventlog.Warning|eventlog.Info)
		if err == nil {
			log, err = eventlog.Open(source)
			if err != nil {
				return nil, fmt.Errorf("Failed to open newly created event log for source '%s'. Inner error: %s", source, err)
			}
		} else {
			return nil, fmt.Errorf("Failed to create open or create existing event log for source '%s'. Inner error: %s", source, err)
		}
	}
	return &eventLogSink{log: log, minLevel: common.INFORMATION}, nil
}

func (s *eventLogSink) Close() error {
	return s.log.Close()
}

func (s *eventLogSink) Write(appScope string, level common.Level, messageTemplate string, fields map[string]interface{}) {
	var msg string
	if appScope == "" {
		msg = common.FormatTemplate(messageTemplate, fields)
	} else {
		msg = fmt.Sprintf("[%s] %s", appScope, common.FormatTemplate(messageTemplate, fields))
	}
	var err error
	for i := 0; i == 0 || i < 5 && err != nil; i++ {
		if level <= common.INFORMATION {
			err = s.log.Info(uint32(1), msg)
		} else if level == common.WARNING {
			err = s.log.Warning(uint32(1), msg)
		} else {
			err = s.log.Error(uint32(1), msg)
		}
	}
	// TODO: Handle this error properly
}
