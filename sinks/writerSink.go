package sinks

import (
	"fmt"
	"io"
	"time"

	"encoding/json"

	"bytes"

	"github.com/jseely/logging/common"
)

type writerSink struct {
	writer       io.WriteCloser
	outputAsBlob bool
}

func NewWriterSink(writer io.WriteCloser, outputAsBlob bool) common.Sink {
	return &writerSink{
		writer:       writer,
		outputAsBlob: outputAsBlob,
	}
}

func (s *writerSink) Close() error {
	return s.writer.Close()
}

func (s *writerSink) Write(appScope string, level common.Level, messageTemplate string, fields map[string]interface{}) {
	if s.outputAsBlob {
		msg := eventHubsMessage{
			Timestamp:        time.Now().UTC().String(),
			ApplicationScope: appScope,
			Level:            level.String(),
			MessageTemplate:  messageTemplate,
			Message:          common.FormatTemplate(messageTemplate, fields),
			Fields:           fields,
		}
		serialized, err := json.Marshal(msg)
		if err != nil {
			return
		}
		var out bytes.Buffer
		json.Indent(&out, serialized, "", "  ")
		s.writer.Write(out.Bytes())
	} else {
		if appScope == "" {
			s.writer.Write([]byte(fmt.Sprintf("%v %s\t%s\n", time.Now(), level.String(), common.FormatTemplate(messageTemplate, fields))))
		} else {
			s.writer.Write([]byte(fmt.Sprintf("%v [%s] %s\t%s\n", time.Now(), appScope, level.String(), common.FormatTemplate(messageTemplate, fields))))
		}
	}
}
