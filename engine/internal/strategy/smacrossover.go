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
	rsiMain     indicators.StatefulIndicator

	currShort   int64
	currLong    int64
	currRsi     int64
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
		rsiMain:     &indicators.RSI{Period: 14}, // Default RSI mapped for oscillator bindings natively
	}
}

// CalculateSignal parses internal ticks asynchronously broadcasting cleanly via EDA bounds
func (s *SMACrossover) CalculateSignal(market *event.MarketEvent, bus *event.EventQueue) {
	currShort, err1 := s.smaShort.Update(market.Bar)
	currLong, err2 := s.smaLong.Update(market.Bar)
	currRsi, err3 := s.rsiMain.Update(market.Bar)

	if err1 != nil || err2 != nil || err3 != nil {
		return
	}

	if !s.isReady {
		s.prevShort = currShort
		s.prevLong = currLong
		s.currRsi = currRsi
		s.isReady = true
		return
	}

	prevShort := s.prevShort
	prevLong := s.prevLong

	s.prevShort = currShort
	s.prevLong = currLong
	s.currShort = currShort
	s.currLong = currLong
	s.currRsi = currRsi

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

// GetIndicators exposes the explicitly instantiated elements mapped
func (s *SMACrossover) GetIndicators() map[string]int64 {
	return map[string]int64{
		"sma_short": s.currShort,
		"sma_long":  s.currLong,
		"rsi_main":  s.currRsi,
	}
}
