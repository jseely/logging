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

var newLoggerTests = []struct {
	logFunc       func(log Logger)
	expectedEvent event
}{
	{logFunc: func(log Logger) { log.Verbose("test1") }, expectedEvent: event{AppScope: "", Level: common.VERBOSE.String(), Message: "test1", Fields: map[string]interface{}{}}},
}

func TestNewLogger(t *testing.T) {
	sink := testSink{events: []event{}}
	log := New(common.VERBOSE, &sink)

	for _, test := range newLoggerTests {
		test.logFunc(log)
		sink.assertContainsEvent(t, test.expectedEvent)
	}
}

func TestLogger(t *testing.T) {
	sink := testSink{events: []event{}}
	log := NewWithApplicationScope("application", common.INFORMATION, &sink)
	log.Information("test")
	sink.assertContainsEvent(t, event{AppScope: "application", Level: common.INFORMATION.String(), Message: "test", Fields: map[string]interface{}{}})

	log.Verbose("test2")
	sink.assertNotContainsEvent(t, event{AppScope: "application", Level: common.VERBOSE.String(), Message: "test2", Fields: map[string]interface{}{}})

	log2 := log.WithApplicationScope("sub", common.VERBOSE)
	log2.Verbose("test3")
	sink.assertContainsEvent(t, event{AppScope: "application.sub", Level: common.VERBOSE.String(), Message: "test3", Fields: map[string]interface{}{}})

	log2.Fatal("test4")
	sink.assertContainsEvent(t, event{AppScope: "application.sub", Level: common.FATAL.String(), Message: "test4", Fields: map[string]interface{}{}})

	log3 := log2.WithApplicationScope("furthersub", common.FATAL)
	log3.Verbose("test5")
	sink.assertNotContainsEvent(t, event{AppScope: "application.sub.furthersub", Level: common.VERBOSE.String(), Message: "test5", Fields: map[string]interface{}{}})

	log3.Error("test6")
	sink.assertNotContainsEvent(t, event{AppScope: "application.sub.furthersub", Level: common.ERROR.String(), Message: "test6", Fields: map[string]interface{}{}})

	log3.Fatal("test7")
	sink.assertContainsEvent(t, event{AppScope: "application.sub.furthersub", Level: common.FATAL.String(), Message: "test7", Fields: map[string]interface{}{}})
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
	AppScope string
	Level    string
	Message  string
	Fields   map[string]interface{}
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
		AppScope: appScope,
		Level:    level.String(),
		Message:  messageTemplate,
		Fields:   fields,
	})
}

func (s *testSink) assertNotContainsEvent(t *testing.T, e event) {
	for _, e1 := range s.events {
		f, _ := json.Marshal(e.Fields)
		f1, _ := json.Marshal(e1.Fields)
		if e.AppScope == e1.AppScope && e.Level == e1.Level && string(f) == string(f1) && e.Message == e1.Message {
			serialized, _ := json.Marshal(s.events)
			es, _ := json.Marshal(e)
			t.Fatalf("Sink did not contain expected event (%s). Sink contents: ", string(es), string(serialized))
		}
	}
}

func (s *testSink) assertContainsEvent(t *testing.T, e event) {
	for _, e1 := range s.events {
		f, _ := json.Marshal(e.Fields)
		f1, _ := json.Marshal(e1.Fields)
		if e.AppScope == e1.AppScope && e.Level == e1.Level && string(f) == string(f1) && e.Message == e1.Message {
			return
		}
	}
	serialized, _ := json.Marshal(s.events)
	es, _ := json.Marshal(e)
	t.Fatalf("Sink did not contain expected event (%s). Sink contents: %s", string(es), string(serialized))
}
