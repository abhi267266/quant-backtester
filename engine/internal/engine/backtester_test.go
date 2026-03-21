package engine

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/logger"
	"github.com/quant-backtester/engine/internal/strategy"
)

// MockDataHandler for TDD
type MockDataHandler struct {
	bars []data.Bar
}

func (m *MockDataHandler) Load() ([]data.Bar, error) { return m.bars, nil }
func (m *MockDataHandler) LoadHead(n int) ([]data.Bar, error) { return m.bars, nil }
func (m *MockDataHandler) LoadTail(n int) ([]data.Bar, error) { return m.bars, nil }
func (m *MockDataHandler) LoadRange(s, e int) ([]data.Bar, error) { return m.bars, nil }
func (m *MockDataHandler) LoadStats() (data.Stats, error) { return data.Stats{}, nil }
func (m *MockDataHandler) Stream(visitor func(b data.Bar, idx int) bool) error {
	for i, b := range m.bars {
		if !visitor(b, i) {
			break
		}
	}
	return nil
}

// MockStrategy
type MockStrategy struct {
	Count int
}

func (m *MockStrategy) OnBar(b data.Bar) strategy.Signal {
	m.Count++
	// Return Buy on 2nd bar to test logging
	if m.Count == 2 {
		return strategy.Signal{Action: strategy.Buy, Price: b.Close}
	}
	return strategy.Signal{Action: strategy.Hold, Price: b.Close}
}

func TestRun(t *testing.T) {
	handler := &MockDataHandler{
		bars: []data.Bar{
			{Timestamp: time.Now(), Close: 100000000},
			{Timestamp: time.Now(), Close: 200000000},
			{Timestamp: time.Now(), Close: 300000000},
		},
	}
	strat := &MockStrategy{}

	err := Run(handler, strat, 10000*data.Decimals, &logger.NoOpLogger{}, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if strat.Count != 3 {
		t.Errorf("expected strategy to process 3 bars, got %d", strat.Count)
	}
}
