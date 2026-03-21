package portfolio

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
)

// Helper to quickly assert equity invariant
func assertInvariant(t *testing.T, p *Portfolio, currentPrice int64) {
	t.Helper()
	expectedEquity := p.InitialCapital + p.RealizedPnL + p.UnrealizedPnL(currentPrice)
	actualEquity := p.GetAccountValue(currentPrice)
	if expectedEquity != actualEquity {
		t.Errorf("Invariant failed! Expected Equity: %d, Actual Equity: %d. (Init: %d, Realized: %d, Unrealized: %d)",
			expectedEquity, actualEquity, p.InitialCapital, p.RealizedPnL, p.UnrealizedPnL(currentPrice))
	}
}

// simulateExecution executes the event pipeline strictly validating queues natively
func simulateExecution(p *Portfolio, direction string, price int64, now time.Time) {
	bus := event.NewEventQueue()
	p.UpdateSignal(&event.SignalEvent{Direction: direction, Price: price, Time: now}, bus)
	if !bus.IsEmpty() {
		order := bus.Pop().(*event.OrderEvent)
		p.UpdateFill(&event.FillEvent{
			Direction: order.Direction,
			Qty:       order.Qty,
			Price:     order.Price,
			Cost:      order.Qty * order.Price,
			Time:      order.Time,
		})
	}
}

func TestInitialState(t *testing.T) {
	initialCash := int64(10000 * data.Decimals)
	p := NewPortfolio(initialCash, nil)

	if p.Cash != initialCash {
		t.Errorf("expected initial cash %d, got %d", initialCash, p.Cash)
	}
	if p.PositionSize != 0 {
		t.Errorf("expected 0 position size, got %d", p.PositionSize)
	}

	assertInvariant(t, p, 100*data.Decimals)
}

func TestBuyExecution(t *testing.T) {
	initialCash := int64(10 * data.Decimals) // 10.00
	p := NewPortfolio(initialCash, nil)
	price := int64(3 * data.Decimals) // 3.00
	now := time.Now()

	simulateExecution(p, "BUY", price, now)

	if p.PositionSize != 3 {
		t.Errorf("expected 3 units, got %d", p.PositionSize)
	}

	expectedCash := int64(1 * data.Decimals) // 1.00 dust
	if p.Cash != expectedCash {
		t.Errorf("expected remaining cash (dust) %d, got %d", expectedCash, p.Cash)
	}

	assertInvariant(t, p, price)
}

func TestSellExecution(t *testing.T) {
	initialCash := int64(10 * data.Decimals)
	p := NewPortfolio(initialCash, nil)
	now := time.Now()

	buyPrice := int64(3 * data.Decimals)
	simulateExecution(p, "BUY", buyPrice, now)

	sellPrice := int64(5 * data.Decimals)
	simulateExecution(p, "SELL", sellPrice, now)

	if p.PositionSize != 0 {
		t.Errorf("expected 0 units after sell, got %d", p.PositionSize)
	}

	expectedCash := int64(16 * data.Decimals)
	if p.Cash != expectedCash {
		t.Errorf("expected %d cash after sell, got %d", expectedCash, p.Cash)
	}

	expectedPnL := int64(6 * data.Decimals)
	if p.RealizedPnL != expectedPnL {
		t.Errorf("expected %d RealizedPnL, got %d", expectedPnL, p.RealizedPnL)
	}

	assertInvariant(t, p, sellPrice)
}

func TestZeroCash(t *testing.T) {
	initialCash := int64(2 * data.Decimals)
	p := NewPortfolio(initialCash, nil)
	now := time.Now()

	price := int64(3 * data.Decimals)
	simulateExecution(p, "BUY", price, now)

	if p.PositionSize != 0 {
		t.Errorf("expected 0 units, got %d", p.PositionSize)
	}
	if p.Cash != initialCash {
		t.Errorf("expected unchanged cash")
	}

	assertInvariant(t, p, price)
}

func TestDrawdownLogic(t *testing.T) {
	initialCash := int64(100 * data.Decimals)
	p := NewPortfolio(initialCash, nil)
	now := time.Now()

	simulateExecution(p, "BUY", 10*data.Decimals, now)

	p.UpdatePrice(20 * data.Decimals)
	if p.PeakEquity != 200*data.Decimals {
		t.Errorf("expected peak equity 200, got %v", p.PeakEquity)
	}

	p.UpdatePrice(10 * data.Decimals)
	if p.MaxDrawdown != 100*data.Decimals {
		t.Errorf("expected max drawdown 100, got %v", p.MaxDrawdown)
	}

	p.UpdatePrice(5 * data.Decimals)
	if p.MaxDrawdown != 150*data.Decimals {
		t.Errorf("expected max drawdown 150, got %v", p.MaxDrawdown)
	}

	assertInvariant(t, p, 5*data.Decimals)
}

func BenchmarkPortfolioProcessing(b *testing.B) {
	initialCash := int64(1000 * data.Decimals)
	p := NewPortfolio(initialCash, nil)
	price := int64(10 * data.Decimals)
	now := time.Now()
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			simulateExecution(p, "BUY", price, now)
		} else {
			simulateExecution(p, "SELL", price+data.Decimals, now)
		}
		p.UpdatePrice(price)
	}
}
