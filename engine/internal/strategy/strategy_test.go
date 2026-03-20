package strategy

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
)

func TestSMACrossover(t *testing.T) {
	strategy := NewSMACrossover(2, 4)

	// Mock price trend
	// data.Bar Scale is 100,000,000
	const scale = 100000000

	tests := []struct {
		name     string
		price    int64
		expected Action
	}{
		// Bar 1: Price 10, SMA2: -, SMA4: - -> Hold
		{"Day 1", 10 * scale, Hold},
		// Bar 2: Price 10, SMA2: 10, SMA4: - -> Hold
		{"Day 2", 10 * scale, Hold},
		// Bar 3: Price 10, SMA2: 10, SMA4: - -> Hold
		{"Day 3", 10 * scale, Hold},
		// Bar 4: Price 10, SMA2: 10, SMA4: 10 -> Hold (No strict crossover)
		{"Day 4", 10 * scale, Hold},
		
		// Bar 5: Price 12, SMA2: 11, SMA4: 10.5 -> Short(11) crosses above Long(10.5) from below(10<=10) -> BUY
		{"Day 5 (Cross Up)", 12 * scale, Buy},
		
		// Bar 6: Price 14, SMA2: 13, SMA4: 11.5 -> Short > Long but didn't cross *on this bar* -> HOLD
		{"Day 6 (Hold Up)", 14 * scale, Hold},
		
		// Bar 7: Price 8, SMA2: 11, SMA4: 11 -> Short == Long -> HOLD (or depending on strict inequality. Let's assume strict cross means Short < Long later)
		{"Day 7 (Equal)", 8 * scale, Hold},

		// Bar 8: Price 6, SMA2: 7, SMA4: 10 -> Short(7) < Long(10) after being >= Long. Crosses below -> SELL
		{"Day 8 (Cross Down)", 6 * scale, Sell},

		// Bar 9: Price 6, SMA2: 6, SMA4: 8.5 -> Short < Long but already crossed -> HOLD
		{"Day 9 (Hold Down)", 6 * scale, Hold},
	}

	ts := time.Now()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bar := data.Bar{
				Timestamp: ts,
				Close:     tc.price,
			}
			ts = ts.Add(24 * time.Hour)

			signal := strategy.OnBar(bar)
			if signal.Action != tc.expected {
				t.Errorf("expected %v, got %v for price %d", tc.expected, signal.Action, tc.price/scale)
			}
		})
	}
}
