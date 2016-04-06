package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
)

type Event struct {
	EventType string
	AppID     string
}

var UnknownEvent = &Event{}

type EventPump struct {
	client marathon.Marathon
	sseCh  marathon.EventsChannel
	Next   chan *Event
}

func newEventPump(client marathon.Marathon) (*EventPump, error) {
	eventCh := make(chan *Event, 5)
	sseCh := make(marathon.EventsChannel, 5)
	if err := client.AddEventsListener(sseCh, marathon.EVENTS_APPLICATIONS); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not subscribe to Marathon application events: %v", err))
	}

	pump := &EventPump{client, sseCh, eventCh}
	go runPump(pump)

	return pump, nil
}

func (p *EventPump) Close() {
	if p.sseCh != nil {
		log.Debug("Closing Marathon SSE channel.")
		p.client.RemoveEventsListener(p.sseCh)
		p.sseCh = nil

		close(p.Next)
	}
}

func runPump(pump *EventPump) {
	for {
		sse, ok := <-pump.sseCh
		if !ok {
			log.Debug("Marathon SSE channel has closed!")
		} else {
			log.Debugf("Marathon SSE event received: %v", sse)

			event := mapEvent(sse)
			if event != UnknownEvent {
				pump.Next <- event
			} else {
				log.Errorf("Unhandled Marathon SSE event: %s", sse.Name)
			}
		}
	}
}

func mapEvent(sse *marathon.Event) *Event {
	switch sse.Name {
	case "status_update_event":
		sue, _ := sse.Event.(*marathon.EventStatusUpdate)
		return &Event{
			EventType: sse.Name,
			AppID:     sue.AppID,
		}
	case "app_terminated_event":
		ate, _ := sse.Event.(*marathon.EventAppTerminated)
		return &Event{
			EventType: sse.Name,
			AppID:     ate.AppID,
		}
	}
	return UnknownEvent
}
