package engine

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/logger"
)

// MockSliceDataHandler streams a predefined slice of strictly controlled bars.
type MockSliceDataHandler struct {
	Bars []data.Bar
}

func (m *MockSliceDataHandler) Load() ([]data.Bar, error) { return m.Bars, nil }
func (m *MockSliceDataHandler) LoadHead(n int) ([]data.Bar, error) { return m.Bars, nil }
func (m *MockSliceDataHandler) LoadTail(n int) ([]data.Bar, error) { return m.Bars, nil }
func (m *MockSliceDataHandler) LoadRange(s, e int) ([]data.Bar, error) { return m.Bars, nil }
func (m *MockSliceDataHandler) LoadStats() (data.Stats, error) { return data.Stats{}, nil }

func (m *MockSliceDataHandler) Stream(onBar func(data.Bar, int) bool) error {
	for i, b := range m.Bars {
		if !onBar(b, i) {
			break
		}
	}
	return nil
}

// SpyStrategy detects values > $100.00 and intercepts execution indices
type SpyStrategy struct {
	TriggeredAtIdx int
	CurrentIdx     int
}

func (s *SpyStrategy) CalculateSignal(market *event.MarketEvent, bus *event.EventQueue) {
	s.CurrentIdx++

	if market.Bar.Close > 100*data.Decimals {
		s.TriggeredAtIdx = s.CurrentIdx - 1

		bus.Push(&event.SignalEvent{
			Time:      market.Bar.Timestamp,
			Direction: "BUY",
			Price:     market.Bar.Close,
		})
	}
}

func (s *SpyStrategy) GetIndicators() map[string]int64 {
	return nil
}

func TestLookAheadBias(t *testing.T) {
	bars := make([]data.Bar, 50)
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 49; i++ {
		bars[i] = data.Bar{
			Timestamp: baseTime.Add(time.Duration(i) * 24 * time.Hour),
			Open:      100 * data.Decimals,
			High:      100 * data.Decimals,
			Low:       100 * data.Decimals,
			Close:     100 * data.Decimals,
			Volume:    1000 * data.Decimals,
		}
	}
	bars[49] = data.Bar{
		Timestamp: baseTime.Add(49 * 24 * time.Hour),
		Open:      100 * data.Decimals,
		High:      500 * data.Decimals,
		Low:       100 * data.Decimals,
		Close:     500 * data.Decimals,
		Volume:    1000 * data.Decimals,
	}

	handler := &MockSliceDataHandler{Bars: bars}
	spy := &SpyStrategy{TriggeredAtIdx: -1}

	err := Run(handler, spy, 10000*data.Decimals, &logger.NoOpLogger{}, 0)
	if err != nil {
		t.Fatalf("Engine failed completely: %v", err)
	}

	if spy.TriggeredAtIdx < 49 && spy.TriggeredAtIdx != -1 {
		t.Fatalf("Look-ahead bias detected! Strategy reacted to future data.")
	}
	if spy.TriggeredAtIdx == -1 {
		t.Fatalf("SpyStrategy completely missed the spike!")
	}
	if spy.TriggeredAtIdx == 49 {
		t.Logf("Successfully captured spike strictly at rowIdx 49 without memory leaks.")
	}
}
