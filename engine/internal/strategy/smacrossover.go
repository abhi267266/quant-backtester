package strategy

import (
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/indicators"
)

// SMACrossover implements a simple moving average crossover strategy
type SMACrossover struct {
	shortPeriod int
	longPeriod  int
	
	smaShort    indicators.StatefulIndicator
	smaLong     indicators.StatefulIndicator

	prevShort   int64
	prevLong    int64
	isReady     bool
}

// NewSMACrossover creates a new SMACrossover strategy
func NewSMACrossover(shortPeriod, longPeriod int) *SMACrossover {
	return &SMACrossover{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		smaShort:    &indicators.SMA{Period: shortPeriod},
		smaLong:     &indicators.SMA{Period: longPeriod},
	}
}

// CalculateSignal parses internal ticks asynchronously broadcasting cleanly via EDA bounds
func (s *SMACrossover) CalculateSignal(market *event.MarketEvent, bus *event.EventQueue) {
	currShort, err1 := s.smaShort.Update(market.Bar)
	currLong, err2 := s.smaLong.Update(market.Bar)

	if err1 != nil || err2 != nil {
		return
	}

	if !s.isReady {
		s.prevShort = currShort
		s.prevLong = currLong
		s.isReady = true
		return
	}

	prevShort := s.prevShort
	prevLong := s.prevLong

	s.prevShort = currShort
	s.prevLong = currLong

	if prevShort <= prevLong && currShort > currLong {
		bus.Push(&event.SignalEvent{
			Time:      market.Bar.Timestamp,
			Direction: "BUY",
			Price:     market.Bar.Close,
		})
		return
	}

	if prevShort >= prevLong && currShort < currLong {
		bus.Push(&event.SignalEvent{
			Time:      market.Bar.Timestamp,
			Direction: "SELL",
			Price:     market.Bar.Close,
		})
		return
	}
}
