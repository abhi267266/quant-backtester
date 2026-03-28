package strategy

import (
	"encoding/json"
	"fmt"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/indicators"
)

// IndicatorConfig maps to a single user-defined indicator in the JSON schema.
type IndicatorConfig struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

// RuleConfig describes a single boolean rule evaluated on each bar.
type RuleConfig struct {
	Type         string   `json:"type"`          // "crossover", "crossunder", "greater_than", "less_than"
	LeftOperand  string   `json:"left_operand"`  // The subject indicator ID
	RightOperand string   `json:"right_operand,omitempty"` // The target indicator ID
	Value        *float64 `json:"value,omitempty"`     // A static boundary threshold
}

type RulesObject struct {
	Buy  []RuleConfig `json:"buy"`
	Sell []RuleConfig `json:"sell"`
}

// DynamicStrategyConfig is the overall JSON envelop structure.
type DynamicStrategyConfig struct {
	StrategyName string            `json:"strategy_name"`
	Indicators   []IndicatorConfig `json:"indicators"`
	Rules        RulesObject       `json:"rules"`
}

// IndicatorConstructor represents a function capable of building a StatefulIndicator from JSON bytes.
type IndicatorConstructor func(params json.RawMessage) (indicators.StatefulIndicator, error)

var indicatorRegistry = map[string]IndicatorConstructor{}

// init registers the base indicators supported by the engine.
func init() {
	indicatorRegistry["SMA"] = func(params json.RawMessage) (indicators.StatefulIndicator, error) {
		var p struct{ Period int `json:"period"` }
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return &indicators.SMA{Period: p.Period}, nil
	}
	indicatorRegistry["EMA"] = func(params json.RawMessage) (indicators.StatefulIndicator, error) {
		var p struct{ Period int `json:"period"` }
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return &indicators.EMA{Period: p.Period}, nil
	}
	indicatorRegistry["RSI"] = func(params json.RawMessage) (indicators.StatefulIndicator, error) {
		var p struct{ Period int `json:"period"` }
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return &indicators.RSI{Period: p.Period}, nil
	}
}

// DynamicStrategy implements the Strategy interface natively.
type DynamicStrategy struct {
	Name       string
	components map[string]indicators.StatefulIndicator
	currValues map[string]int64
	prevValues map[string]int64
	buyRules   []RuleConfig
	sellRules  []RuleConfig
	isReady    bool
}

// NewDynamicStrategyFromJSON parses a raw JSON configuration payload and initializes all mapped indicators.
func NewDynamicStrategyFromJSON(configBytes []byte) (*DynamicStrategy, error) {
	var cfg DynamicStrategyConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON strategy: %w", err)
	}

	ds := &DynamicStrategy{
		Name:       cfg.StrategyName,
		components: make(map[string]indicators.StatefulIndicator),
		currValues: make(map[string]int64),
		prevValues: make(map[string]int64),
		buyRules:   cfg.Rules.Buy,
		sellRules:  cfg.Rules.Sell,
	}

	for _, ic := range cfg.Indicators {
		factory, exists := indicatorRegistry[ic.Type]
		if !exists {
			return nil, fmt.Errorf("unknown indicator type requested: %s", ic.Type)
		}

		indicator, err := factory(ic.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to build indicator %s: %w", ic.ID, err)
		}
		ds.components[ic.ID] = indicator
	}

	return ds, nil
}

// evaluateRule processes a single RuleConfig mathematically.
func (ds *DynamicStrategy) evaluateRule(rule RuleConfig) bool {
	primCurr := ds.currValues[rule.LeftOperand]
	primPrev := ds.prevValues[rule.LeftOperand]

	var secCurr, secPrev int64
	if rule.RightOperand != "" {
		secCurr = ds.currValues[rule.RightOperand]
		secPrev = ds.prevValues[rule.RightOperand]
	} else if rule.Value != nil {
		scaledLimit := int64(*rule.Value * float64(data.Decimals))
		secCurr = scaledLimit
		secPrev = scaledLimit
	} else {
		return false
	}

	switch rule.Type {
	case "crossover":
		return (primPrev <= secPrev) && (primCurr > secCurr)
	case "crossunder":
		return (primPrev >= secPrev) && (primCurr < secCurr)
	case "greater_than":
		return (primCurr > secCurr)
	case "less_than":
		return (primCurr < secCurr)
	}
	return false
}

// CalculateSignal complies rigidly with the EDA contract, iteratively executing rule evaluations per tick.
func (ds *DynamicStrategy) CalculateSignal(market *event.MarketEvent, bus *event.EventQueue) {
	hasInsufficientData := false

	for id, indicator := range ds.components {
		val, err := indicator.Update(market.Bar)
		if err != nil {
			hasInsufficientData = true
		} else {
			ds.currValues[id] = val
		}
	}

	if hasInsufficientData {
		return
	}

	if !ds.isReady {
		for id, val := range ds.currValues {
			ds.prevValues[id] = val
		}
		ds.isReady = true
		return
	}

	// Evaluate BUY rules (Logical AND)
	buyTriggered := len(ds.buyRules) > 0
	for _, rule := range ds.buyRules {
		if !ds.evaluateRule(rule) {
			buyTriggered = false
			break
		}
	}

	// Evaluate SELL rules (Logical AND)
	sellTriggered := len(ds.sellRules) > 0
	for _, rule := range ds.sellRules {
		if !ds.evaluateRule(rule) {
			sellTriggered = false
			break
		}
	}

	if buyTriggered {
		bus.Push(&event.SignalEvent{
			Time:      market.Bar.Timestamp,
			Direction: "BUY",
			Price:     market.Bar.Close,
		})
	} else if sellTriggered {
		bus.Push(&event.SignalEvent{
			Time:      market.Bar.Timestamp,
			Direction: "SELL",
			Price:     market.Bar.Close,
		})
	}

	// Cycle values implicitly mapping bounded arrays
	for id, val := range ds.currValues {
		ds.prevValues[id] = val
	}
}

// GetIndicators returns the dynamically bound native indicators correctly 
func (ds *DynamicStrategy) GetIndicators() map[string]int64 {
	return ds.currValues
}
