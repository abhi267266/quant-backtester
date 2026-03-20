package strategy

import (
	"github.com/quant-backtester/engine/data"
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

// OnBar evaluates the new bar and returns a trading signal
func (s *SMACrossover) OnBar(bar data.Bar) Signal {
	currShort, err1 := s.smaShort.Update(bar)
	currLong, err2 := s.smaLong.Update(bar)

	// If we don't have enough data for the long SMA, we must hold
	if err1 != nil || err2 != nil {
		return Signal{Action: Hold, Price: bar.Close}
	}

	// The first valid long SMA value serves as the initial previous state
	if !s.isReady {
		s.prevShort = currShort
		s.prevLong = currLong
		s.isReady = true
		return Signal{Action: Hold, Price: bar.Close}
	}

	prevShort := s.prevShort
	prevLong := s.prevLong

	s.prevShort = currShort
	s.prevLong = currLong

	// Cross from below to above -> Buy
	if prevShort <= prevLong && currShort > currLong {
		return Signal{Action: Buy, Price: bar.Close}
	}

	// Cross from above to below -> Sell
	if prevShort >= prevLong && currShort < currLong {
		return Signal{Action: Sell, Price: bar.Close}
	}

	return Signal{Action: Hold, Price: bar.Close}
}
