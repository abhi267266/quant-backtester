package portfolio

import (
	"testing"
	"time"

	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/strategy"
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

	p.ProcessSignal(strategy.Signal{Action: strategy.Buy, Price: price}, now, price)

	// qty = 10 / 3 = 3 units.
	// Cost = 3 * 3 = 9. Remaining Cash = 1.

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
	p.ProcessSignal(strategy.Signal{Action: strategy.Buy, Price: buyPrice}, now, buyPrice)

	// Fast forward to price = 5
	sellPrice := int64(5 * data.Decimals)
	p.ProcessSignal(strategy.Signal{Action: strategy.Sell, Price: sellPrice}, now, sellPrice)

	if p.PositionSize != 0 {
		t.Errorf("expected 0 units after sell, got %d", p.PositionSize)
	}

	// units = 3
	// sell proceeds = 3 * 5 = 15
	// new cash = 1 + 15 = 16
	// Realized PnL = (5 - 3) * 3 = 6
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
	p.ProcessSignal(strategy.Signal{Action: strategy.Buy, Price: price}, now, price)

	// Cannot buy, cash < price
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

	p.ProcessSignal(strategy.Signal{Action: strategy.Buy, Price: 10 * data.Decimals}, now, 10*data.Decimals)
	// Bought 10 units at 10. Cash = 0. Equity = 100.

	// Price goes up to 20. Equity = 200. Peak = 200.
	p.UpdatePrice(20 * data.Decimals)

	if p.PeakEquity != 200*data.Decimals {
		t.Errorf("expected peak equity 200, got %v", p.PeakEquity)
	}

	// Price gaps down to 10. Equity = 100. Peak = 200.
	// Drawdown = Peak - Equity = 100.
	p.UpdatePrice(10 * data.Decimals)
	if p.MaxDrawdown != 100*data.Decimals {
		t.Errorf("expected max drawdown 100, got %v", p.MaxDrawdown)
	}

	// Price drops to 5. Equity = 50. Peak = 200.
	// Drawdown = 150.
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
	sigBuy := strategy.Signal{Action: strategy.Buy, Price: price}
	sigSell := strategy.Signal{Action: strategy.Sell, Price: price + data.Decimals}
	now := time.Now()
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			p.ProcessSignal(sigBuy, now, price)
		} else {
			p.ProcessSignal(sigSell, now, price+data.Decimals)
		}
		p.UpdatePrice(price)
	}
}
