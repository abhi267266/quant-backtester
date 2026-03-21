package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/portfolio"
	"github.com/quant-backtester/engine/internal/strategy"
)

// formatScaledPrice returns a human-readable string for our 10^8 scaled integer
func formatScaledPrice(val int64) string {
	return fmt.Sprintf("%.8f", float64(val)/float64(data.Decimals))
}

// Run executes the given strategy streamingly over the DataHandler internally
func Run(handler data.DataHandler, s strategy.Strategy, initialCash int64) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Timestamp", "Action", "Price")

	hasSignals := false
	port := portfolio.NewPortfolio(initialCash)
	var lastClose int64

	err := handler.Stream(func(b data.Bar, rowIdx int) bool {
		lastClose = b.Close
		
		signal := s.OnBar(b)
		port.ProcessSignal(signal, b.Close)
		port.UpdatePrice(b.Close)

		if signal.Action != strategy.Hold {
			hasSignals = true
			table.Append(
				b.Timestamp.Format(time.RFC3339),
				string(signal.Action),
				formatScaledPrice(signal.Price),
			)
		}
		return true // continue streaming
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
