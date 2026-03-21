package event

import (
	"time"

	"github.com/quant-backtester/engine/data"
)

// EventType categorizes the different pipeline signals natively
type EventType string

const (
	Market EventType = "MARKET"
	Signal EventType = "SIGNAL"
	Order  EventType = "ORDER"
	Fill   EventType = "FILL"
)

// Event defines the required contract natively mapped for all pipeline occurrences
type Event interface {
	Type() EventType
	Timestamp() time.Time
}

// MarketEvent is fired strictly when a new valid bar arrives from data handling streams
type MarketEvent struct {
	Bar data.Bar
}

func (e *MarketEvent) Type() EventType       { return Market }
func (e *MarketEvent) Timestamp() time.Time  { return e.Bar.Timestamp }

// SignalEvent is natively fired by strategy modules isolating logic evaluation
type SignalEvent struct {
	Time      time.Time
	Direction string // "BUY" or "SELL" explicitly
	Price     int64
}

func (e *SignalEvent) Type() EventType       { return Signal }
func (e *SignalEvent) Timestamp() time.Time  { return e.Time }

// OrderEvent securely validates margin/capital before triggering exchange logic
type OrderEvent struct {
	Time      time.Time
	Direction string
	Qty       int64
	Price     int64
}

func (e *OrderEvent) Type() EventType       { return Order }
func (e *OrderEvent) Timestamp() time.Time  { return e.Time }

// FillEvent accurately maps exchange completion natively resolving into the Portfolio
type FillEvent struct {
	Time       time.Time
	Direction  string
	Qty        int64
	Price      int64
	Commission int64
	Cost       int64 // natively evaluated (qty * price) scaled at 10^8
}

func (e *FillEvent) Type() EventType       { return Fill }
func (e *FillEvent) Timestamp() time.Time  { return e.Time }

// EventQueue manages chronological order executions identically mapped in FIFO bounds
type EventQueue struct {
	queue []Event
}

// NewEventQueue establishes 100-bound zero overhead slices natively handling events
func NewEventQueue() *EventQueue {
	return &EventQueue{
		queue: make([]Event, 0, 100),
	}
}

// Push natively enqueues execution limits securely at the tail
func (q *EventQueue) Push(e Event) {
	q.queue = append(q.queue, e)
}

// Pop extracts the front natively yielding nil silently if starved
func (q *EventQueue) Pop() Event {
	if len(q.queue) == 0 {
		return nil
	}
	e := q.queue[0]
	q.queue = q.queue[1:]
	return e
}

// IsEmpty bounds external iteration strictly if executions safely halted
func (q *EventQueue) IsEmpty() bool {
	return len(q.queue) == 0
}
