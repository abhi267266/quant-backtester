package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/quant-backtester/engine/data"
)

// formatScaledPrice returns a human-readable string for our 10^8 scaled integer
func formatScaledPrice(val int64) string {
	return fmt.Sprintf("%.8f", float64(val)/float64(data.Decimals))
}

func printBarsTable(bars []data.Bar) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Timestamp", "Open", "High", "Low", "Close", "Volume")
	
	for _, b := range bars {
		table.Append(
			b.Timestamp.Format(time.RFC3339),
			formatScaledPrice(b.Open),
			formatScaledPrice(b.High),
			formatScaledPrice(b.Low),
			formatScaledPrice(b.Close),
			formatScaledPrice(b.Volume),
		)
	}
	table.Render()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'inspect' subcommand")
		os.Exit(1)
	}

	if os.Args[1] != "inspect" {
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("expected inspect subcommand: 'head', 'tail', 'range', or 'stats'")
		os.Exit(1)
	}
	
	subcommand := os.Args[2]
	
	headCmd := flag.NewFlagSet("head", flag.ExitOnError)
	headN := headCmd.Int("n", 10, "Number of rows to display")

	tailCmd := flag.NewFlagSet("tail", flag.ExitOnError)
	tailN := tailCmd.Int("n", 10, "Number of rows to display")

	rangeCmd := flag.NewFlagSet("range", flag.ExitOnError)
	rangeStart := rangeCmd.Int("start", 0, "Start row index (inclusive)")
	rangeEnd := rangeCmd.Int("end", 10, "End row index (exclusive)")

	statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)

	handler := &data.CSVDataHandler{Reader: os.Stdin}

	switch subcommand {
	case "head":
		headCmd.Parse(os.Args[3:])
		bars, err := handler.LoadHead(*headN)
		if err != nil {
			log.Fatalf("failed to load head data: %v", err)
		}
		printBarsTable(bars)

	case "tail":
		tailCmd.Parse(os.Args[3:])
		bars, err := handler.LoadTail(*tailN)
		if err != nil {
			log.Fatalf("failed to load tail data: %v", err)
		}
		printBarsTable(bars)

	case "range":
		rangeCmd.Parse(os.Args[3:])
		bars, err := handler.LoadRange(*rangeStart, *rangeEnd)
		if err != nil {
			log.Fatalf("failed to load range data: %v", err)
		}
		printBarsTable(bars)

	case "stats":
		statsCmd.Parse(os.Args[3:])
		stats, err := handler.LoadStats()
		if err != nil {
			log.Fatalf("failed to load stats: %v", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Metric", "Value")
		table.Append("Total Bars", stats.Count)
		table.Append("Start Range", stats.Start.Format(time.RFC3339))
		table.Append("End Range", stats.End.Format(time.RFC3339))
		
		status := "OK"
		if stats.MissingPeriod {
			status = "WARNING (Time skips detected)"
		}
		table.Append("Continuity Check", status)
		table.Render()

	default:
		fmt.Printf("unknown inspect subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
