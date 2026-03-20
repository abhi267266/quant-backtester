package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/strategy"
)

// formatScaledPrice returns a human-readable string for our 10^8 scaled integer
func formatScaledPrice(val int64) string {
	return fmt.Sprintf("%.8f", float64(val)/float64(data.Decimals))
}

// Run executes the given strategy streamingly over the DataHandler internally
func Run(handler data.DataHandler, s strategy.Strategy) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Timestamp", "Action", "Price")

	hasSignals := false

	err := handler.Stream(func(b data.Bar, rowIdx int) bool {
		signal := s.OnBar(b)

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

	return nil
}
