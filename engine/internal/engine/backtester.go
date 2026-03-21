package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/event"
	"github.com/quant-backtester/engine/internal/execution"
	"github.com/quant-backtester/engine/internal/logger"
	"github.com/quant-backtester/engine/internal/portfolio"
	"github.com/quant-backtester/engine/internal/strategy"
)

// formatScaledPrice returns a human-readable string for our 10^8 scaled integer
func formatScaledPrice(val int64) string {
	return fmt.Sprintf("%.8f", float64(val)/float64(data.Decimals))
}

// Run uniquely navigates backtests natively iterating over infinite Event-Driven pipelines safely avoiding procedural coupling
func Run(handler data.DataHandler, s strategy.Strategy, initialCash int64, l logger.LogWriter, snapshotInterval int) error {
	defer l.Flush()
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Timestamp", "Action", "Price", "Capital", "Cash", "CostBasis", "PositionSize")

	hasSignals := false
	port := portfolio.NewPortfolio(initialCash, l)
	execHandler := execution.NewExecutionHandler()
	var lastClose int64
	bus := event.NewEventQueue()

	err := handler.Stream(func(b data.Bar, rowIdx int) bool {
		lastClose = b.Close

		// Initialize pipeline sequence converting Raw Stream to an Asynchronous Event Native Tick Array
		bus.Push(&event.MarketEvent{Bar: b})

		// Maintain sequence entirely blocking execution ticks identically navigating across state limits securely
		for !bus.IsEmpty() {
			currEvent := bus.Pop()

			switch e := currEvent.(type) {
			case *event.MarketEvent:
				port.UpdatePrice(e.Bar.Close) // Persist native accounting value updates identically per tick internally limits isolated
				s.CalculateSignal(e, bus)
			case *event.SignalEvent:
				port.UpdateSignal(e, bus)
			case *event.OrderEvent:
				execHandler.Execute(e, bus)
			case *event.FillEvent:
				port.UpdateFill(e)
				
				hasSignals = true
				table.Append(
					e.Time.Format(time.RFC3339),
					e.Direction,
					formatScaledPrice(e.Price),
					formatScaledPrice(port.GetAccountValue(e.Price)),
					formatScaledPrice(port.Cash),
					formatScaledPrice(port.CostBasis),
					fmt.Sprintf("%d", port.PositionSize),
				)
			}
		}

		if snapshotInterval > 0 && rowIdx%snapshotInterval == 0 {
			l.LogSnapshot(logger.SnapshotEntry{
				Timestamp:     b.Timestamp,
				TotalEquity:   port.GetAccountValue(b.Close),
				Cash:          port.Cash,
				UnrealizedPnL: port.UnrealizedPnL(b.Close),
			})
		}
		return true // trigger consecutive iteration flawlessly routing data streams 
	})

	if err != nil {
		return err
	}

	if hasSignals {
		fmt.Println("\n--- Backtest Signals ---")
		table.Render()
	} else {
		fmt.Println("\n--- Backtest Complete (No Trading Signals) ---")
	}

	finalEquity := port.GetAccountValue(lastClose)
	netProfit := finalEquity - port.InitialCapital

	fmt.Println("\n--- Performance Summary ---")
	summary := tablewriter.NewWriter(os.Stdout)
	summary.Header("Metric", "Value")
	summary.Append("Initial Capital", formatScaledPrice(port.InitialCapital))
	summary.Append("Final Equity", formatScaledPrice(finalEquity))
	summary.Append("Net Profit", formatScaledPrice(netProfit))
	summary.Append("Max Drawdown", formatScaledPrice(port.MaxDrawdown))
	summary.Render()

	return nil
}
