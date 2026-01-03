package handylib

import (
	"sync"
	"time"
)

type Event[T any] struct {
	Payload   T
	Timestamp time.Time
}

func NewEvent[T any](payload T) Event[T] {
	return Event[T]{
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

type EventStore[TPayload any, TState any] struct {
	mu           sync.RWMutex
	events       []Event[TPayload]
	reducer      func(state TState, event Event[TPayload]) TState
	initialState TState
}

func NewEventStore[TPayload any, TState any](initialState TState, reducer func(state TState, event Event[TPayload]) TState) EventStore[TPayload, TState] {
	return EventStore[TPayload, TState]{
		initialState: initialState,
		reducer:      reducer,
	}
}

func (e *EventStore[TPayload, TState]) GetEvents() []Event[TPayload] {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.events
}

func (e *EventStore[TPayload, TState]) AppendEvent(event Event[TPayload]) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, event)
}

func (e *EventStore[TPayload, TState]) BuildState() TState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// TODO: if TState is a map or slice, the following line will cause the modification of `e.initialState`
	// possible fixes: reducer must copy and return a new state; require an initial state factory and a copy function for state on EventStore initialization
	state := e.initialState
	for _, event := range e.events {
		state = e.reducer(state, event)
	}

	return state
}
