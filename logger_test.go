package logging

import (
	"encoding/json"
	"testing"

	"github.com/jseely/logging/common"
)

func TestCreateFieldsMap(t *testing.T) {
	fields := createFieldsMap("", []interface{}{})
	if len(fields) != 0 {
		t.Fatalf("Fields should be empty")
	}

	assertPanic(t, "Too few values provided", func() { createFieldsMap("{s}", []interface{}{}) })

	fields = createFieldsMap("{s}", []interface{}{"test"})
	if len(fields) != 1 || fields["s"].(string) != "test" {
		exp, _ := json.Marshal(map[string]interface{}{"s": "test"})
		act, _ := json.Marshal(fields)
		t.Fatalf("Fields did not contain the correct number of entries or the value \"test\". Expected: %s Actual: %s", string(exp), string(act))
	}

	fields = createFieldsMap("{s} {v} {s}", []interface{}{"test1", "test2"})
	if len(fields) != 2 || fields["s"].(string) != "test1" || fields["v"].(string) != "test2" {
		exp, _ := json.Marshal(map[string]interface{}{"s": "test1", "v": "test2"})
		act, _ := json.Marshal(fields)
		t.Fatalf("Fields did not contain the correct number of entries or the value \"test\". Expected: %s Actual: %s", string(exp), string(act))
	}
}

func TestLogger(t *testing.T) {
	sink := testSink{events: []event{}}
	log := New(common.INFORMATION, &sink)
	log.Information("test")
	sink.assertContainsEvent(t, event{level: common.INFORMATION, message: "test", fields: map[string]interface{}{}})

	log.Verbose("test2")
	sink.assertNotContainsEvent(t, event{level: common.VERBOSE, message: "test2", fields: map[string]interface{}{}})
}

func assertPanic(t *testing.T, failMessage string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf(failMessage)
		}
	}()
	f()
}

type event struct {
	appScope string
	level    common.Level
	message  string
	fields   map[string]interface{}
}

type testSink struct {
	events []event
}

func (s *testSink) Close() error {
	return nil
}

func (s *testSink) Level(minLevel common.Level) {}

func (s *testSink) Write(appScope string, level common.Level, messageTemplate string, fields map[string]interface{}) {
	s.events = append(s.events, event{
		appScope: appScope,
		level:    level,
		message:  messageTemplate,
		fields:   fields,
	})
}

func (s *testSink) assertNotContainsEvent(t *testing.T, e event) {
	for _, e1 := range s.events {
		f, _ := json.Marshal(e.fields)
		f1, _ := json.Marshal(e1.fields)
		if e.appScope == e1.appScope && e.level == e1.level && string(f) == string(f1) && e.message == e1.message {
			serialized, _ := json.Marshal(s.events)
			es, _ := json.Marshal(e)
			t.Fatalf("Sink did not contain expected event (%s). Sink contents: ", string(es), string(serialized))
		}
	}
}

func (s *testSink) assertContainsEvent(t *testing.T, e event) {
	for _, e1 := range s.events {
		f, _ := json.Marshal(e.fields)
		f1, _ := json.Marshal(e1.fields)
		if e.appScope == e1.appScope && e.level == e1.level && string(f) == string(f1) && e.message == e1.message {
			return
		}
	}
	serialized, _ := json.Marshal(s.events)
	es, _ := json.Marshal(e)
	t.Fatalf("Sink did not contain expected event (%s). Sink contents: ", string(es), string(serialized))
}
