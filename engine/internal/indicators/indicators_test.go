package indicators

import (
	"reflect"
	"testing"

	"github.com/quant-backtester/engine/data"
)

func TestIndicators(t *testing.T) {
	// Sample data using the 10^8 Scale
	// Bar 1: 10 * 10^8
	// Bar 2: 12 * 10^8
	// Bar 3: 15 * 10^8
	// Bar 4: 14 * 10^8
	// Bar 5: 18 * 10^8
	bars := []data.Bar{
		{Close: 1000000000},
		{Close: 1200000000},
		{Close: 1500000000},
		{Close: 1400000000},
		{Close: 1800000000},
	}

	tests := []struct {
		name      string
		indicator BatchIndicator
		expected  []int64
		wantErr   bool
	}{
		{
			name:      "SMA Period 3",
			indicator: &SMA{Period: 3},
			expected:  []int64{0, 0, 1233333333, 1366666666, 1566666666},
			wantErr:   false,
		},
		{
			name:      "SMA Insufficient Data",
			indicator: &SMA{Period: 10},
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "EMA Period 3",
			indicator: &EMA{Period: 3},
			// EMA init is SMA of first 3 prices: 1233333333
			// next string price=14: EMA = 1233333333 + (1400000000 - 1233333333) * 2 / 4 = 1316666666
			// next string price=18: EMA = 1316666666 + (1800000000 - 1316666666) * 2 / 4 = 1558333333
			expected: []int64{0, 0, 1233333333, 1316666666, 1558333333},
			wantErr:  false,
		},
		{
			name:      "EMA Insufficient Data",
			indicator: &EMA{Period: 10},
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "RSI Period 3",
			indicator: &RSI{Period: 3},
			// RSI period 3 requires 3 differences (4 bars).
			// differences: 
			// 1 -> 2: gain +2
			// 2 -> 3: gain +3
			// 3 -> 4: loss -1
			// sum_gain_3 = 5
			// sum_loss_3 = 1
			// rsi_3 = 100 * 5 / 6 = 83.33333333 = 8333333333
			// 4 -> 5: gain +4
			// sum_gain_4 = 5 - 5/3 + 4 = 5 - 1.66666666 + + 4
			// 500000000 - 166666666 + 400000000 = 733333334
			// sum_loss_4 = 100000000 - 33333333 + 0 = 66666667
			// rsi_4 = 100 * 10^8 * 733333334 / (733333334 + 66666667) = 10000000000 * 733333334 / 800000001
			// 7333333340000000000 / 800000001 = 9166666663
			expected: []int64{0, 0, 0, 8333333333, 9166666663},
			wantErr:  false,
		},
		{
			name:      "RSI Insufficient Data",
			indicator: &RSI{Period: 10},
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.indicator.Compute(bars)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error %v, got err %v", tt.wantErr, err)
				return
			}
			if err == nil {
				// verify length matches input
				if len(got) != len(bars) {
					t.Errorf("expected slice length %d, got %d", len(bars), len(got))
				}
				// verify precision to nearest integer
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("expected %v\ngot      %v", tt.expected, got)
				}
			}
		})
	}
}

func BenchmarkIndicators(b *testing.B) {
	bars := make([]data.Bar, 1000)
	for i := 0; i < 1000; i++ {
		bars[i] = data.Bar{Close: int64(1000000000 + i*10000)}
	}

	sma := &SMA{Period: 14}
	b.Run("SMA", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = sma.Compute(bars)
		}
	})

	ema := &EMA{Period: 14}
	b.Run("EMA", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = ema.Compute(bars)
		}
	})

	rsi := &RSI{Period: 14}
	b.Run("RSI", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = rsi.Compute(bars)
		}
	})
}
