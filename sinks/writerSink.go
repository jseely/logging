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
	minLevel     common.Level
	writer       io.WriteCloser
	outputAsBlob bool
}

func NewWriterSink(writer io.WriteCloser, outputAsBlob bool) common.Sink {
	return &writerSink{
		minLevel:     common.INFORMATION,
		writer:       writer,
		outputAsBlob: outputAsBlob,
	}
}

func (s *writerSink) Close() error {
	return s.writer.Close()
}

func (s *writerSink) Level(level common.Level) {
	s.minLevel = level
}

func (s *writerSink) Write(level common.Level, messageTemplate string, fields map[string]interface{}) {
	if level < s.minLevel {
		return
	}
	if s.outputAsBlob {
		msg := eventHubsMessage{
			Timestamp:       time.Now().UTC().String(),
			Level:           level.String(),
			MessageTemplate: messageTemplate,
			Message:         common.FormatTemplate(messageTemplate, fields),
			Fields:          fields,
		}
		serialized, err := json.Marshal(msg)
		if err != nil {
			return
		}
		var out bytes.Buffer
		json.Indent(&out, serialized, "", "  ")
		s.writer.Write(out.Bytes())
	} else {
		s.writer.Write([]byte(fmt.Sprintf("%v %s\t%s\n", time.Now(), level.String(), common.FormatTemplate(messageTemplate, fields))))
	}
}
