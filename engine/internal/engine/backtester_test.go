package engine

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/logger"
)

// MockDataHandler for TDD
type MockDataHandler struct{}

func (m *MockDataHandler) Load() ([]data.Bar, error) { return nil, nil }
func (m *MockDataHandler) LoadHead(n int) ([]data.Bar, error) { return nil, nil }
func (m *MockDataHandler) LoadTail(n int) ([]data.Bar, error) { return nil, nil }
func (m *MockDataHandler) LoadRange(s, e int) ([]data.Bar, error) { return nil, nil }
func (m *MockDataHandler) LoadStats() (data.Stats, error) { return data.Stats{}, nil }

func (m *MockDataHandler) Stream(visitor func(b data.Bar, rowIdx int) bool) error {
	bar := data.Bar{
		Timestamp: time.Date(2026, 3, 21, 14, 49, 52, 0, time.UTC), // Sample Time
		Close:     2 * data.Decimals,
	}
	visitor(bar, 0)
	return nil
}

// MockStrategy always returns Buy signals natively mapping EDA events
type MockStrategy struct{}

func (s *MockStrategy) CalculateSignal(market *event.MarketEvent, bus *event.EventQueue) {
	bus.Push(&event.SignalEvent{
		Time:      market.Bar.Timestamp,
		Direction: "BUY",
		Price:     market.Bar.Close,
	})
}

func TestRun(t *testing.T) {
	handler := &MockDataHandler{}
	strat := &MockStrategy{}

	l := &logger.NoOpLogger{}

	err := Run(handler, strat, 10000*data.Decimals, l, 0)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
}
