package indicators

import (
	"errors"
	"math/bits"

	"github.com/quant-backtester/engine/data"
)

const Scale int64 = 100000000

var ErrInsufficientData = errors.New("insufficient data for indicator period")

type Indicator interface {
	Compute(bars []data.Bar) ([]int64, error)
}

// SMA defines a Simple Moving Average indicator
type SMA struct {
	Period int
}

func (s *SMA) Compute(bars []data.Bar) ([]int64, error) {
	if len(bars) < s.Period {
		return nil, ErrInsufficientData
	}

	result := make([]int64, len(bars))
	var sum int64

	for i := 0; i < len(bars); i++ {
		sum += bars[i].Close
		if i >= s.Period {
			sum -= bars[i-s.Period].Close
		}
		if i >= s.Period-1 {
			result[i] = sum / int64(s.Period)
		}
	}

	return result, nil
}

// EMA defines an Exponential Moving Average indicator
type EMA struct {
	Period int
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

// RSI defines a Relative Strength Index indicator
type RSI struct {
	Period int
}

func (r *RSI) Compute(bars []data.Bar) ([]int64, error) {
	// RSI needs Period + 1 bars minimum exactly because it measures Period differences
	if len(bars) < r.Period+1 {
		return nil, ErrInsufficientData
	}

	result := make([]int64, len(bars))
	var sumGain, sumLoss int64

	// Calculate initial simple sum of gain/loss
	for i := 1; i <= r.Period; i++ {
		diff := bars[i].Close - bars[i-1].Close
		if diff > 0 {
			sumGain += diff
		} else if diff < 0 {
			sumLoss -= diff
		}
	}

	for i := r.Period; i < len(bars); i++ {
		if i > r.Period {
			diff := bars[i].Close - bars[i-1].Close
			var gain, loss int64
			if diff > 0 {
				gain = diff
			} else if diff < 0 {
				loss = -diff
			}
			// Wilder's Smoothing method logic using sum
			// sum = prev_sum - (prev_sum / period) + current
			sumGain = sumGain - (sumGain / int64(r.Period)) + gain
			sumLoss = sumLoss - (sumLoss / int64(r.Period)) + loss
		}

		if sumLoss == 0 {
			result[i] = 100 * Scale
		} else if sumGain == 0 {
			result[i] = 0
		} else {
			// Compute: (100 * Scale * sumGain) / (sumGain + sumLoss) using 128-bit math
			multiplier := uint64(100 * Scale)
			hi, lo := bits.Mul64(multiplier, uint64(sumGain))
			quo, _ := bits.Div64(hi, lo, uint64(sumGain+sumLoss))
			
			result[i] = int64(quo)
		}
	}

	return result, nil
}
