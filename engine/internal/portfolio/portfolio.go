package portfolio

import (
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/logger"
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
	logger         logger.LogWriter
}

// NewPortfolio initializes an empty portfolio
func NewPortfolio(initialCash int64, l logger.LogWriter) *Portfolio {
	if l == nil {
		l = &logger.NoOpLogger{}
	}
	return &Portfolio{
		InitialCapital: initialCash,
		Cash:           initialCash,
		PeakEquity:     initialCash,
		logger:         l,
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

// UpdateSignal validates a SignalEvent natively evaluating cash constraints to push an OrderEvent securely
func (p *Portfolio) UpdateSignal(sig *event.SignalEvent, bus *event.EventQueue) {
	if sig.Direction == "BUY" {
		if p.PositionSize == 0 && p.Cash >= sig.Price {
			qty := p.Cash / sig.Price
			if qty > 0 {
				bus.Push(&event.OrderEvent{
					Time:      sig.Time,
					Direction: "BUY",
					Qty:       qty,
					Price:     sig.Price,
				})
			}
		}
	} else if sig.Direction == "SELL" {
		if p.PositionSize > 0 {
			bus.Push(&event.OrderEvent{
				Time:      sig.Time,
				Direction: "SELL",
				Qty:       p.PositionSize,
				Price:     sig.Price,
			})
		}
	}
}

// UpdateFill rigorously manages exchange broker transactions into accounting states seamlessly integrating scaled bounds
func (p *Portfolio) UpdateFill(fill *event.FillEvent) {
	if fill.Direction == "BUY" {
		p.PositionSize += fill.Qty
		p.CostBasis += fill.Cost
		p.Cash -= (fill.Cost + fill.Commission)

		p.logger.LogTrade(logger.TradeEntry{
			Timestamp:  fill.Time,
			Side:       "BUY",
			Price:      fill.Price,
			Qty:        fill.Qty,
			TotalValue: fill.Cost,
		})
	} else if fill.Direction == "SELL" {
		proceeds := fill.Cost
		pnl := proceeds - p.CostBasis
		p.RealizedPnL += pnl

		p.logger.LogTrade(logger.TradeEntry{
			Timestamp:  fill.Time,
			Side:       "SELL",
			Price:      fill.Price,
			Qty:        fill.Qty,
			TotalValue: proceeds,
		})

		p.Cash += (proceeds - fill.Commission)
		p.PositionSize -= fill.Qty
		if p.PositionSize == 0 {
			p.CostBasis = 0
		}
	}
}
