package indicators

import "github.com/quant-backtester/engine/data"

// SMA defines a Simple Moving Average indicator
type SMA struct {
	Period int

	// internal state
	history []int64
	sum     int64
}

func (s *SMA) Update(bar data.Bar) (int64, error) {
	s.history = append(s.history, bar.Close)
	s.sum += bar.Close

	if len(s.history) > s.Period {
		s.sum -= s.history[0]
		s.history = s.history[1:]
	}

	if len(s.history) < s.Period {
		return 0, ErrInsufficientData
	}

	return s.sum / int64(s.Period), nil
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
