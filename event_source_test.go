package handylib

import (
	"fmt"
	"maps"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestOperationType int

const (
	ITEM_ADDED TestOperationType = iota
	ITEM_REMOVED
)

type TestPayload struct {
	ID   int
	Type TestOperationType
}

type TestState map[int]int

// reducer does not modify the state; instead, it copies and returns a new instance of state
var reducer = func(state TestState, event Event[TestPayload]) TestState {
	newState := make(TestState)
	maps.Copy(newState, state)

	switch event.Payload.Type {
	case ITEM_ADDED:
		newState[event.Payload.ID] = newState[event.Payload.ID] + 1
	case ITEM_REMOVED:
		if qty, ok := newState[event.Payload.ID]; ok && qty > 0 {
			newState[event.Payload.ID] = qty - 1
		}
	default:
		panic(fmt.Sprintf("unknown operation type: %v", event.Payload.Type))
	}
	return newState
}

func TestEventStoreAppend(t *testing.T) {
	e := NewEventStore[TestPayload, TestState](make(map[int]int), reducer)

	assert.Empty(t, e.GetEvents())

	event := NewEvent(TestPayload{
		ID:   1,
		Type: ITEM_ADDED,
	})

	e.AppendEvent(event)

	assert.Equal(t, 1, len(e.GetEvents()))
	assert.Equal(t, event, e.GetEvents()[0])
}

func TestEventStoreAppendConcurrent(t *testing.T) {
	e := NewEventStore[TestPayload, TestState](make(map[int]int), reducer)

	assert.Empty(t, e.GetEvents())

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			e.AppendEvent(NewEvent(TestPayload{
				ID:   i,
				Type: ITEM_ADDED,
			}))
		})
	}
	wg.Wait()

	assert.Equal(t, 10, len(e.GetEvents()))
}

func TestReplayEvents(t *testing.T) {
	e := NewEventStore[TestPayload, TestState](make(map[int]int), reducer)
	e.AppendEvent(NewEvent(TestPayload{ID: 1, Type: ITEM_ADDED}))
	e.AppendEvent(NewEvent(TestPayload{ID: 2, Type: ITEM_ADDED}))

	quantities := e.BuildState()

	assert.Equal(t, 1, quantities[1])
	assert.Equal(t, 1, quantities[2])

	e.AppendEvent(NewEvent(TestPayload{ID: 1, Type: ITEM_ADDED}))

	quantities = e.BuildState()

	assert.Equal(t, 2, quantities[1])
	assert.Equal(t, 1, quantities[2])

	e.AppendEvent(NewEvent(TestPayload{ID: 3, Type: ITEM_ADDED}))

	quantities = e.BuildState()

	assert.Equal(t, 2, quantities[1])
	assert.Equal(t, 1, quantities[2])
	assert.Equal(t, 1, quantities[3])

	e.AppendEvent(NewEvent(TestPayload{ID: 1, Type: ITEM_REMOVED}))

	quantities = e.BuildState()

	assert.Equal(t, 1, quantities[1])
	assert.Equal(t, 1, quantities[2])
	assert.Equal(t, 1, quantities[3])
}
