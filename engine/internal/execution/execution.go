package execution

import (
	"github.com/quant-backtester/engine/internal/event"
)

// ExecutionHandler mocks exchange mechanics validating Orders seamlessly
type ExecutionHandler struct{}

// NewExecutionHandler binds the simulator locally
func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{}
}

// Execute parses an OrderEvent and constructs a deterministically evaluated FillEvent
func (ex *ExecutionHandler) Execute(order *event.OrderEvent, bus *event.EventQueue) {
	fill := &event.FillEvent{
		Time:       order.Time,
		Direction:  order.Direction,
		Qty:        order.Qty,
		Price:      order.Price,
		Commission: 0,
		Cost:       order.Qty * order.Price,
	}
	bus.Push(fill)
}
