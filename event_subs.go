package main

import "sync"

func NewEventSubs() *EventSubs {
	return &EventSubs{
		Subs: make(map[interface{}]chan *Event),
	}
}

type EventSubs struct {
	Subs map[interface{}]chan *Event
	mu   sync.RWMutex
}

func (es *EventSubs) Subscribe(subscriber interface{}) chan *Event {
	ch := make(chan *Event)

	es.mu.Lock()
	es.Subs[subscriber] = ch
	es.mu.Unlock()

	return ch
}

func (es *EventSubs) Unsubscribe(subscriber interface{}) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if ch, exists := es.Subs[subscriber]; exists {
		close(ch)
		delete(es.Subs, subscriber)
	}
}

func (es *EventSubs) Publish(event *Event) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, ch := range es.Subs {
		select {
		case ch <- event:
		default:
		}
	}
}
