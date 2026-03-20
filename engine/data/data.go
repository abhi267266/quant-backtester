package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Bar represents a single unit of OHLCV data
type Bar struct {
	Timestamp time.Time `json:"timestamp"`
	Open      int64     `json:"open"`
	High      int64     `json:"high"`
	Low       int64     `json:"low"`
	Close     int64     `json:"close"`
	Volume    int64     `json:"volume"`
}

// Stats contains metadata about the dataset, including continuity checks
type Stats struct {
	Count         int
	Start         time.Time
	End           time.Time
	MissingPeriod bool
}

// DataHandler defines an interface for data loading and streaming
type DataHandler interface {
	Load() ([]Bar, error)
	LoadHead(n int) ([]Bar, error)
	LoadTail(n int) ([]Bar, error)
	LoadRange(start, end int) ([]Bar, error)
	LoadStats() (Stats, error)
	Stream(visitor func(b Bar, rowIdx int) bool) error
}

// CSVDataHandler implements DataHandler for CSV sources
type CSVDataHandler struct {
	Reader io.Reader
}

const (
	// Decimals specifies the scaling factor for string to integer conversion
	// 8 decimals supports down to saturating satoshis (100,000,000)
	Decimals int64 = 100000000
)

// parseDecimal safely converts string to scaled int64 (s * Decimals)
func parseDecimal(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty decimal string")
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid decimal string: %s", s)
	}

	intPartStr := parts[0]
	var intPart int64
	var err error

	if intPartStr != "" && intPartStr != "-" {
		intPart, err = strconv.ParseInt(intPartStr, 10, 64)
		if err != nil {
			return 0, err
		}
	} else if intPartStr == "-" {
		intPart = -0 // Just needed to indicate negative sign applied later if int_part was zero
	} // if "" then it's 0

	var fracPartStr string
	if len(parts) == 2 {
		fracPartStr = parts[1]
	}

	// Pad with zeros up to 8 places
	if len(fracPartStr) > 8 {
		fracPartStr = fracPartStr[:8] // truncate to 8 places
	}
	for len(fracPartStr) < 8 {
		fracPartStr += "0"
	}

	var fracPart int64
	if fracPartStr != "00000000" {
		fracPart, err = strconv.ParseInt(fracPartStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}

	isNegative := strings.HasPrefix(s, "-")
	
	result := intPart * Decimals
	if isNegative {
		result -= fracPart
	} else {
		result += fracPart
	}

	return result, nil
}

// parseTimestamp handles both RFC3339 and YYYY-MM-DD formats
func parseTimestamp(ts string) (time.Time, error) {
	// Try RFC3339 first
	parsed, err := time.Parse(time.RFC3339, ts)
	if err == nil {
		return parsed, nil
	}
	// Fallback to YYYY-MM-DD
	return time.Parse("2006-01-02", ts)
}

// Stream iterates over the CSV sequentially, parsing Bars without retaining them in memory.
// Yields parsed Bar and 0-indexed data row number (excluding header) to the visitor.
// Return false from visitor to stop streaming.
func (h *CSVDataHandler) Stream(visitor func(b Bar, rowIdx int) bool) error {
	reader := csv.NewReader(h.Reader)
	rowIdx := 0

	for i := 0; ; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read csv data at line %d: %w", i+1, err)
		}

		if i == 0 {
			continue // skip header
		}

		if len(record) < 6 {
			return fmt.Errorf("row %d: incomplete data, expected 6 columns got %d", i+1, len(record))
		}

		timestamp, err := parseTimestamp(record[0])
		if err != nil {
			return fmt.Errorf("row %d: invalid timestamp: %w", i+1, err)
		}

		open, err := parseDecimal(record[1])
		if err != nil {
			return fmt.Errorf("row %d: invalid open: %w", i+1, err)
		}

		high, err := parseDecimal(record[2])
		if err != nil {
			return fmt.Errorf("row %d: invalid high: %w", i+1, err)
		}

		low, err := parseDecimal(record[3])
		if err != nil {
			return fmt.Errorf("row %d: invalid low: %w", i+1, err)
		}

		closeVal, err := parseDecimal(record[4])
		if err != nil {
			return fmt.Errorf("row %d: invalid close: %w", i+1, err)
		}

		volume, err := parseDecimal(record[5])
		if err != nil {
			return fmt.Errorf("row %d: invalid volume: %w", i+1, err)
		}

		bar := Bar{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeVal,
			Volume:    volume,
		}

		if cont := visitor(bar, rowIdx); !cont {
			break
		}
		rowIdx++
	}

	return nil
}

// Load reads and parses the entire underlying source into a slice of Bars
func (h *CSVDataHandler) Load() ([]Bar, error) {
	var bars []Bar
	err := h.Stream(func(b Bar, _ int) bool {
		bars = append(bars, b)
		return true
	})
	return bars, err
}

// LoadHead reads only the first n records
func (h *CSVDataHandler) LoadHead(n int) ([]Bar, error) {
	var bars []Bar
	err := h.Stream(func(b Bar, rowIdx int) bool {
		if rowIdx >= n {
			return false
		}
		bars = append(bars, b)
		return true
	})
	return bars, err
}

// LoadTail reads the entire file but retains only a rolling buffer of size n, costing bounded O(n) RAM
func (h *CSVDataHandler) LoadTail(n int) ([]Bar, error) {
	if n <= 0 {
		return nil, nil
	}
	buffer := make([]Bar, 0, n)
	err := h.Stream(func(b Bar, rowIdx int) bool {
		if len(buffer) < n {
			buffer = append(buffer, b)
		} else {
			// shift left and append
			copy(buffer, buffer[1:])
			buffer[n-1] = b
		}
		return true
	})
	return buffer, err
}

// LoadRange yields rows where index is in [start, end)
func (h *CSVDataHandler) LoadRange(start, end int) ([]Bar, error) {
	var bars []Bar
	err := h.Stream(func(b Bar, rowIdx int) bool {
		if rowIdx >= end {
			return false
		}
		if rowIdx >= start {
			bars = append(bars, b)
		}
		return true
	})
	return bars, err
}

// LoadStats computes aggregate statistics in a stream format with O(1) memory
func (h *CSVDataHandler) LoadStats() (Stats, error) {
	var stats Stats
	var lastTime time.Time
	var expectedInterval time.Duration

	err := h.Stream(func(b Bar, rowIdx int) bool {
		if rowIdx == 0 {
			stats.Start = b.Timestamp
		} else if rowIdx == 1 {
			expectedInterval = b.Timestamp.Sub(lastTime)
		} else if rowIdx > 1 {
			if !stats.MissingPeriod {
				diff := b.Timestamp.Sub(lastTime)
				if diff != expectedInterval {
					stats.MissingPeriod = true
				}
			}
		}

		stats.End = b.Timestamp
		stats.Count++
		lastTime = b.Timestamp
		return true
	})

	return stats, err
}
