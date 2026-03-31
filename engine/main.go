package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
	"github.com/quant-backtester/engine/data"
	"github.com/quant-backtester/engine/internal/engine"
	"github.com/quant-backtester/engine/internal/indicators"
	"github.com/quant-backtester/engine/internal/logger"
	"github.com/quant-backtester/engine/internal/strategy"
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

func printIndicatorTable(bars []data.Bar, values []int64, name string, tailN int) {
	if tailN > len(bars) {
		tailN = len(bars)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Timestamp", "Close", name)

	startIdx := len(bars) - tailN
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(bars); i++ {
		table.Append(
			bars[i].Timestamp.Format(time.RFC3339),
			formatScaledPrice(bars[i].Close),
			formatScaledPrice(values[i]),
		)
	}
	table.Render()
}

func main() {
	_ = godotenv.Load() // optional, so we ignore error if file not found
	if len(os.Args) < 2 {
		fmt.Println("expected a subcommand: 'inspect', 'sma', 'ema', 'rsi', 'backtest'")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	var handler data.DataHandler = &data.CSVDataHandler{Reader: os.Stdin}

	switch subcommand {
	case "inspect":
		if len(os.Args) < 3 {
			fmt.Println("expected inspect subcommand: 'head', 'tail', 'range', or 'stats'")
			os.Exit(1)
		}

		inspectType := os.Args[2]

		headCmd := flag.NewFlagSet("head", flag.ExitOnError)
		headN := headCmd.Int("n", 10, "Number of rows to display")

		tailCmd := flag.NewFlagSet("tail", flag.ExitOnError)
		tailN := tailCmd.Int("n", 10, "Number of rows to display")

		rangeCmd := flag.NewFlagSet("range", flag.ExitOnError)
		rangeStart := rangeCmd.Int("start", 0, "Start row index (inclusive)")
		rangeEnd := rangeCmd.Int("end", 10, "End row index (exclusive)")

		statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)

		switch inspectType {
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
			fmt.Printf("unknown inspect subcommand: %s\n", inspectType)
			os.Exit(1)
		}
	case "sma", "ema", "rsi":
		cmd := flag.NewFlagSet(subcommand, flag.ExitOnError)
		period := cmd.Int("period", 14, "Indicator period")
		n := cmd.Int("n", 10, "Number of latest rows to display")

		cmd.Parse(os.Args[2:])

		bars, err := handler.Load()
		if err != nil {
			log.Fatalf("failed to load data: %v", err)
		}

		var ind indicators.BatchIndicator
		var name string
		switch subcommand {
		case "sma":
			ind = &indicators.SMA{Period: *period}
			name = fmt.Sprintf("SMA(%d)", *period)
		case "ema":
			ind = &indicators.EMA{Period: *period}
			name = fmt.Sprintf("EMA(%d)", *period)
		case "rsi":
			ind = &indicators.RSI{Period: *period}
			name = fmt.Sprintf("RSI(%d)", *period)
		}

		results, err := ind.Compute(bars)
		if err != nil {
			log.Fatalf("failed to compute %s: %v", name, err)
		}

		printIndicatorTable(bars, results, name, *n)

	case "backtest":
		cmd := flag.NewFlagSet(subcommand, flag.ExitOnError)
		capital := cmd.Float64("capital", 10000.0, "Initial capital (in standard currency)")
		logFile := cmd.String("log", "", "File to output CSV logs")
		interval := cmd.Int("interval", 100, "Interval for equity snapshots")
		configFile := cmd.String("config", "", "Path to the JSON strategy configuration file")
		mode := cmd.String("mode", "csv", "Data ingestion mode ('csv', 'live', 'api', or 'yfinance')")
		symbol := cmd.String("symbol", "", "Symbol to trade if in live/api/yfinance mode")
		startStr := cmd.String("start", "", "Start date for API backtest (YYYY-MM-DD)")
		endStr := cmd.String("end", "", "End date for API backtest (YYYY-MM-DD)")

		cmd.Parse(os.Args[2:])

		var startDate, endDate time.Time
		if *startStr != "" {
			var parseErr error
			startDate, parseErr = time.Parse("2006-01-02", *startStr)
			if parseErr != nil {
				log.Fatalf("invalid start date format. expected YYYY-MM-DD: %v", parseErr)
			}
		}
		if *endStr != "" {
			var parseErr error
			endDate, parseErr = time.Parse("2006-01-02", *endStr)
			if parseErr != nil {
				log.Fatalf("invalid end date format. expected YYYY-MM-DD: %v", parseErr)
			}
		}

		if *mode == "yfinance" {
			if *symbol == "" {
				log.Fatalf("symbol is required when using yfinance mode. Use -symbol <ticker>")
			}
			if startDate.IsZero() || endDate.IsZero() {
				log.Fatalf("start and end dates are specifically required for yfinance to construct bounds cleanly. Use -start and -end")
			}
			handler = &data.YFinanceDataHandler{
				Symbol:    *symbol,
				StartDate: startDate,
				EndDate:   endDate,
			}
		} else if *mode == "live" || *mode == "api" {
			if *symbol == "" {
				log.Fatalf("symbol is required when using %s mode. Use -symbol <ticker>", *mode)
			}
			apiKey := os.Getenv("ALPHA_API_KEY")
			if apiKey == "" {
				log.Fatalf("ALPHA_API_KEY environment variable not set")
			}
			handler = &data.AlphaVantageDataHandler{
				Symbol:         *symbol,
				APIKey:         apiKey,
				StartDate:      startDate,
				EndDate:        endDate,
				DisablePolling: (*mode == "api"),
			}
		}

		if *configFile == "" {
			log.Fatalf("config file path is required. Use -config <path>")
		}

		configBytes, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("failed to read strategy configuration: %v", err)
		}

		strat, err := strategy.NewDynamicStrategyFromJSON(configBytes)
		if err != nil {
			log.Fatalf("failed to load dynamic strategy: %v", err)
		}

		initialCash := int64(*capital * float64(data.Decimals))

		var l logger.LogWriter = &logger.NoOpLogger{}
		if *logFile != "" {
			file, err := os.Create(*logFile)
			if err != nil {
				log.Fatalf("failed to create log file: %v", err)
			}
			defer file.Close()
			l = logger.NewCSVLogger(file)
		}

		fmt.Printf("Starting backtest with Dynamic JSON Strategy (%s, Initial Capital: %.2f)...\n", strat.Name, *capital)

		err = engine.Run(handler, strat, initialCash, l, *interval)
		if err != nil {
			log.Fatalf("backtest failure: %v", err)
		}

	default:
		fmt.Printf("unknown command: %s\n", subcommand)
		os.Exit(1)
	}
}
