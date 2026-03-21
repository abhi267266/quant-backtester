package strategy

import (
	"github.com/quant-backtester/engine/internal/event"
)

// Strategy formally subscribes to MarketEvents natively broadcasting logic onto the EventQueues
type Strategy interface {
	CalculateSignal(market *event.MarketEvent, bus *event.EventQueue)
}



