package indicators

import "github.com/quant-backtester/engine/data"

// EMA defines an Exponential Moving Average indicator
type EMA struct {
	Period int

	// internal state
	count   int
	sum     int64
	prevEMA int64
}

func (e *EMA) Update(bar data.Bar) (int64, error) {
	e.count++

	if e.count <= e.Period {
		e.sum += bar.Close
		if e.count == e.Period {
			e.prevEMA = e.sum / int64(e.Period)
			return e.prevEMA, nil
		}
		return 0, ErrInsufficientData
	}

	divisor := int64(e.Period + 1)
	currEMA := e.prevEMA + (bar.Close-e.prevEMA)*2/divisor
	e.prevEMA = currEMA
	return currEMA, nil
}

func (e *EMA) Compute(bars []data.Bar) ([]int64, error) {
	if len(bars) < e.Period {
		return nil, ErrInsufficientData
	}

	result := make([]int64, len(bars))
	var sum int64

	// Initial SMA value for the first EMA point
	for i := 0; i < e.Period; i++ {
		sum += bars[i].Close
	}
	result[e.Period-1] = sum / int64(e.Period)
	
	divisor := int64(e.Period + 1)

	for i := e.Period; i < len(bars); i++ {
		prevEMA := result[i-1]
		price := bars[i].Close
		
		// EMA_i = EMA_{i-1} + (Price - EMA_{i-1}) * 2 / (Period + 1)
		result[i] = prevEMA + (price-prevEMA)*2/divisor
	}

	return result, nil
}
