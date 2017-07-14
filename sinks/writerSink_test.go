package sinks

import (
	"bufio"
	"bytes"
	"testing"

	"strings"

	"github.com/jseely/logging"
)

func TestWriterSinkWrite(t *testing.T) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	sink := NewWriterSink(writer, false)

	sink.Write(logging.VERBOSE, "test", map[string]interface{}{})
	sink.Write(logging.DEBUG, "test", map[string]interface{}{})
	writer.Flush()
	if buf.String() != "" {
		t.Fatalf("Default min level should be Information, therefore buffer should be empty.")
	}

	sink.Write(logging.INFORMATION, "test1", map[string]interface{}{})
	sink.Write(logging.WARNING, "test2", map[string]interface{}{})
	sink.Write(logging.ERROR, "test3", map[string]interface{}{})
	sink.Write(logging.FATAL, "test4", map[string]interface{}{})
	writer.Flush()
	if !stringContains(buf.String(), "Information", "test1", "Warning", "test2", "Error", "test3", "Fatal", "test4") {
		t.Fatalf("Default min level should be Information, therefore buffer should contain all levels above and their corresponding messages.\n%s", buf.String())
	}

	buf.Reset()
	sink.Write(logging.INFORMATION, "{s} {v} {s}", map[string]interface{}{"s": "sval", "v": "vval"})
	writer.Flush()
	if !strings.Contains(buf.String(), "sval vval sval") {
		t.Fatalf("Field values should be substituted into the messageTemplate. Expected (sval vval sval) received (%s)", buf.String())
	}
}

func stringContains(str string, values ...string) bool {
	for _, val := range values {
		if !strings.Contains(str, val) {
			return false
		}
	}
	return true
}
