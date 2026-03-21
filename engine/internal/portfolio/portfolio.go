package portfolio

import (
	"github.com/quant-backtester/engine/internal/strategy"
)

// Portfolio manages state with O(1) zero-allocation logic
type Portfolio struct {
	InitialCapital int64
	Cash           int64
	PositionSize   int64 // whole units
	CostBasis      int64 // total cost
	RealizedPnL    int64
	PeakEquity     int64
	MaxDrawdown    int64 // Absolute difference
}

// NewPortfolio initializes an empty portfolio
func NewPortfolio(initialCash int64) *Portfolio {
	return &Portfolio{
		InitialCapital: initialCash,
		Cash:           initialCash,
		PeakEquity:     initialCash,
	}
}

// GetAccountValue calculates current tracking equity
func (p *Portfolio) GetAccountValue(currentPrice int64) int64 {
	return p.Cash + (p.PositionSize * currentPrice)
}

// UnrealizedPnL calculates the open positions PnL
func (p *Portfolio) UnrealizedPnL(currentPrice int64) int64 {
	if p.PositionSize == 0 {
		return 0
	}
	return (p.PositionSize * currentPrice) - p.CostBasis
}

// UpdatePrice modifies the highest peak and max drawdown dynamically per bar
func (p *Portfolio) UpdatePrice(currentPrice int64) {
	equity := p.GetAccountValue(currentPrice)
	if equity > p.PeakEquity {
		p.PeakEquity = equity
	}

	drawdown := p.PeakEquity - equity
	if drawdown > p.MaxDrawdown {
		p.MaxDrawdown = drawdown
	}
}

// ProcessSignal processes a signal and adjusts the portfolio balance and position accordingly
func (p *Portfolio) ProcessSignal(sig strategy.Signal, price int64) {
	switch sig.Action {
	case strategy.Buy:
		if p.Cash >= price {
			qty := p.Cash / price // Whole units
			if qty > 0 {
				cost := qty * price
				p.PositionSize += qty
				p.CostBasis += cost
				p.Cash -= cost
			}
		}
	case strategy.Sell:
		if p.PositionSize > 0 {
			proceeds := p.PositionSize * price

			// Calculate realized PnL for this sale
			pnl := proceeds - p.CostBasis
			p.RealizedPnL += pnl

			p.Cash += proceeds
			p.PositionSize = 0
			p.CostBasis = 0
		}
	case strategy.Hold:
		// do nothing
	}
}
