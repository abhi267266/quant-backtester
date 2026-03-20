package indicators

import (
	"math/rand"
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
)

func generateMockBars(n int) []data.Bar {
	bars := make([]data.Bar, n)
	// Use deterministic seed for repeatable benchmarks
	r := rand.New(rand.NewSource(42))
	basePrice := int64(100 * Scale)

	for i := 0; i < n; i++ {
		// Random price variation
		change := (r.Int63() % (2 * Scale)) - Scale
		basePrice += change
		bars[i] = data.Bar{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Close:     basePrice,
		}
	}
	return bars
}

func BenchmarkSMA_Batch(b *testing.B) {
	bars := generateMockBars(10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate the unoptimized strategy backtest loop
		var history []data.Bar
		for j := 0; j < len(bars); j++ {
			history = append(history, bars[j])
			sma := &SMA{Period: 14}
			// Compute on the entire growing slice
			_, _ = sma.Compute(history)
		}
	}
}

func BenchmarkSMA_Stateful(b *testing.B) {
	bars := generateMockBars(10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate the optimized O(1) strategy backtest loop
		sma := &SMA{Period: 14}
		for j := 0; j < len(bars); j++ {
			// Pass only the single new bar to the StatefulIndicator
			_, _ = sma.Update(bars[j])
		}
	}
}
