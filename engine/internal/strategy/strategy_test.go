package strategy

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
)

// makeBar creates a uniform bar where OHLCV are all the identical price.
func makeBar(price int64, t time.Time) data.Bar {
	return data.Bar{
		Timestamp: t,
		Open:      price,
		High:      price,
		Low:       price,
		Close:     price,
		Volume:    100 * data.Decimals,
	}
}

func TestDynamicStrategyValidation(t *testing.T) {
	var configJSON = []byte(`{
		"strategy_name": "MovingAverageOscillator",
		"indicators": [
			{ "id": "fast_ma", "type": "SMA", "params": { "period": 2 } },
			{ "id": "slow_ma", "type": "SMA", "params": { "period": 4 } },
			{ "id": "rsi_main", "type": "RSI", "params": { "period": 5 } }
		],
		"rules": {
			"buy": [
				{ "type": "crossover", "left_operand": "fast_ma", "right_operand": "slow_ma" }
			],
			"sell": [
				{ "type": "crossunder", "left_operand": "fast_ma", "right_operand": "slow_ma" }
			]
		}
	}`)

	strat, err := NewDynamicStrategyFromJSON(configJSON)
	if err != nil {
		t.Fatalf("failed to initialize strategy natively: %v", err)
	}

	var _ Strategy = strat

	if strat.Name != "MovingAverageOscillator" {
		t.Errorf("expected name MovingAverageOscillator, got %s", strat.Name)
	}
	if len(strat.components) != 3 {
		t.Errorf("expected 3 indicator components, got %d", len(strat.components))
	}
	if len(strat.buyRules)+len(strat.sellRules) != 2 {
		t.Errorf("expected 2 fully loaded rules mapped, got %d", len(strat.buyRules)+len(strat.sellRules))
	}
}

func TestCalculateSignal_EvaluationLogic(t *testing.T) {
	var configJSON = []byte(`{
		"strategy_name": "BasicCross",
		"indicators": [
			{ "id": "fast", "type": "SMA", "params": { "period": 2 } },
			{ "id": "slow", "type": "SMA", "params": { "period": 3 } }
		],
		"rules": {
			"buy": [
				{ "type": "crossover", "left_operand": "fast", "right_operand": "slow" }
			],
			"sell": []
		}
	}`)

	strat, err := NewDynamicStrategyFromJSON(configJSON)
	if err != nil {
		t.Fatalf("failed parsing: %v", err)
	}

	bus := event.NewEventQueue()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	scale := data.Decimals

	// Bar 1 & 2: Prices 10. Indicators cannot evaluate.
	strat.CalculateSignal(&event.MarketEvent{Bar: makeBar(10*scale, baseTime)}, bus)
	if !bus.IsEmpty() { t.Fatal("Signal triggered during warmup") }
	strat.CalculateSignal(&event.MarketEvent{Bar: makeBar(10*scale, baseTime.Add(24*time.Hour))}, bus)
	if !bus.IsEmpty() { t.Fatal("Signal triggered during warmup") }

	// Bar 3: Price 10. (Slow SMA finally has 3 data points).
	// Strategy transitions state to isReady=true. NO signals possible yet because no Delta (prev == nil)
	strat.CalculateSignal(&event.MarketEvent{Bar: makeBar(10*scale, baseTime.Add(48*time.Hour))}, bus)
	if !bus.IsEmpty() { t.Fatal("expected no signal on exact warmup tick") }
    if !strat.isReady { t.Fatal("expected isReady state flag identically switched") }

	// Bar 4: Price 10. Fast=10, Slow=10. (Prev was 10,10). No event.
	strat.CalculateSignal(&event.MarketEvent{Bar: makeBar(10*scale, baseTime.Add(72*time.Hour))}, bus)
	if !bus.IsEmpty() { t.Fatal("expected no signal, holds flat natively") }

	// Bar 5: Price 12. 
	// Fast SMA (Period 2) = (10+12)/2 = 11.
	// Slow SMA (Period 3) = (10+10+12)/3 = 10.66.
	// Primary (11) > Secondary (10.66). Prev (10 <= 10). Crossover valid!
	strat.CalculateSignal(&event.MarketEvent{Bar: makeBar(12*scale, baseTime.Add(96*time.Hour))}, bus)
	
	if bus.IsEmpty() {
		t.Fatal("expected crossover signal triggered identically bounded")
	}

	sig := bus.Pop().(*event.SignalEvent)
	if sig.Direction != "BUY" {
		t.Fatalf("expected BUY, got %s", sig.Direction)
	}
}
