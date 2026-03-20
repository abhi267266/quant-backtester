package indicators

import (
	"math/bits"

	"github.com/quant-backtester/engine/data"
)

// RSI defines a Relative Strength Index indicator
type RSI struct {
	Period int

	// internal state
	history []int64
	count   int
	sumGain int64
	sumLoss int64
}

func (r *RSI) Update(bar data.Bar) (int64, error) {
	r.history = append(r.history, bar.Close)
	if len(r.history) > 2 {
		r.history = r.history[1:]
	}

	r.count++

	if r.count == 1 {
		return 0, ErrInsufficientData
	}

	diff := bar.Close - r.history[0]
	var gain, loss int64
	if diff > 0 {
		gain = diff
	} else if diff < 0 {
		loss = -diff
	}

	if r.count <= r.Period+1 {
		r.sumGain += gain
		r.sumLoss += loss

		if r.count < r.Period+1 {
			return 0, ErrInsufficientData
		}
	} else {
		r.sumGain = r.sumGain - (r.sumGain / int64(r.Period)) + gain
		r.sumLoss = r.sumLoss - (r.sumLoss / int64(r.Period)) + loss
	}

	if r.sumLoss == 0 {
		return 100 * Scale, nil
	} else if r.sumGain == 0 {
		return 0, nil
	}

	multiplier := uint64(100 * Scale)
	hi, lo := bits.Mul64(multiplier, uint64(r.sumGain))
	quo, _ := bits.Div64(hi, lo, uint64(r.sumGain+r.sumLoss))
	
	return int64(quo), nil
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
