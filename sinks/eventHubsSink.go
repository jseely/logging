package sinks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"fmt"

	"github.com/jseely/logging/common"
)

type eventHubsSink struct {
	minLevel      common.Level
	enrichWith    map[string]interface{}
	uri           string
	auth          string
	wg            *sync.WaitGroup
	eventBuffer   []interface{}
	bufferMut     *sync.Mutex
	terminate     chan bool
	batchMessages bool
}

type eventHubsBatchMessage struct {
	Body           string
	UserProperties map[string]string
}

type eventHubsMessage struct {
	Timestamp        string                 `json:"@timestamp"`
	ApplicationScope string                 `json:"appScope"`
	Level            string                 `json:"level"`
	MessageTemplate  string                 `json:"messageTemplate"`
	Message          string                 `json:"message"`
	Fields           map[string]interface{} `json:"fields"`
}

type EventHubsSink interface {
	common.Sink
	EnrichWith(key string, value interface{})
}

func NewEventHubsSink(uri, auth string) EventHubsSink {
	sink := &eventHubsSink{
		minLevel:      common.INFORMATION,
		enrichWith:    map[string]interface{}{},
		uri:           uri,
		auth:          auth,
		wg:            &sync.WaitGroup{},
		eventBuffer:   make([]interface{}, 0, 50),
		bufferMut:     &sync.Mutex{},
		terminate:     make(chan bool),
		batchMessages: false,
	}
	sink.wg.Add(1)
	go sink.eventSender()
	return sink
}

func (s *eventHubsSink) Close() error {
	s.terminate <- true
	s.wg.Wait()
	return nil
}

func (s *eventHubsSink) eventSender() {
	defer s.wg.Done()
	tick := time.Tick(time.Second * 10)
	for {
		select {
		case <-s.terminate:
			return
		case <-tick:
			if len(s.eventBuffer) == 0 {
				continue
			}
			s.bufferMut.Lock()
			eventSet := s.eventBuffer
			s.eventBuffer = make([]interface{}, 0, 50)
			s.bufferMut.Unlock()

			if s.batchMessages {
				err := sendEventHubBatch(s.uri, s.auth, eventSet)
				if err != nil {
					log.Println(err.Error())
				} else {
					log.Printf("Successfully sent batch of %v messages to event hubs", len(eventSet))
				}
			} else {
				for _, message := range eventSet {
					msg, _ := message.(eventHubsMessage)
					err := sendEventHubMessage(s.uri, s.auth, msg)
					if err != nil {
						log.Println(err.Error())
					} else {
						log.Println("Successfully send message to event hubs.")
					}
				}
			}
		}
	}
}

func sendEventHubMessage(uri, auth string, message eventHubsMessage) error {
	serialized, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Failed to serialize event hubs message. %s", err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(serialized))
	if err != nil {
		return fmt.Errorf("Failed to create request. %s", err.Error())
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/atom+xml;type=entry;charset=utf-8")
	req.Header.Set("Type", "SerilogEvent")
	req.Header.Set("Timestamp", message.Timestamp)
	req.Header.Set("MessageId", uuid.NewV4().String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request. %s", err.Error())
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("Request did not complete successfully, returned '%s'", resp.Status)
	}
	return nil
}

func sendEventHubBatch(uri, auth string, batch []interface{}) error {
	serialized, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("Failed to serialize batched events. %s", err.Error())
	}
	log.Println(string(serialized))
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(serialized))
	if err != nil {
		return fmt.Errorf("Failed to create request. %s", err.Error())
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/vnd.microsoft.servicebus.json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request. %s", err.Error())
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("Request did not complete successfully, returned '%s'", resp.Status)
	}
	return nil
}

func (s *eventHubsSink) Level(minLevel common.Level) {
	s.minLevel = minLevel
}

func (s *eventHubsSink) EnrichWith(key string, value interface{}) {
	s.enrichWith[key] = value
}

func (s *eventHubsSink) Write(appScope string, level common.Level, messageTemplate string, fields map[string]interface{}) {
	for k, v := range s.enrichWith {
		fields[k] = v
	}
	event := eventHubsMessage{
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		ApplicationScope: appScope,
		Level:            level.String(),
		MessageTemplate:  messageTemplate,
		Message:          common.FormatTemplate(messageTemplate, fields),
		Fields:           fields,
	}
	if s.batchMessages {
		serialized, err := json.Marshal(event)
		if err != nil {
			log.Println("Failed to serialize event")
			return
		}
		evtBufMsg := eventHubsBatchMessage{
			Body: string(serialized),
			UserProperties: map[string]string{
				"Type":      "SerilogEvent",
				"Timestamp": event.Timestamp,
				"MessageId": uuid.NewV4().String(),
			},
		}
		s.bufferMut.Lock()
		s.eventBuffer = append(s.eventBuffer, evtBufMsg)
		s.bufferMut.Unlock()
	} else {
		s.bufferMut.Lock()
		s.eventBuffer = append(s.eventBuffer, event)
		s.bufferMut.Unlock()
	}
}
