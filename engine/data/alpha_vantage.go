package data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

// AlphaVantageDataHandler manages fetching historical data via REST and natively polling for continuous live ticks
type AlphaVantageDataHandler struct {
	Symbol       string
	APIKey       string
	EndpointURL  string        // Used mainly to mock in tests; falls back to standard if empty
	PollInterval time.Duration // Interval to sleep between live polls
	MaxPollCount int           // Primarily for testing termination. -1 or 0 for infinite live polling.
}

type avResponse struct {
	TimeSeries map[string]map[string]string `json:"Time Series (1min)"`
}

func (h *AlphaVantageDataHandler) fetchAndParse(url string) ([]Bar, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alpha vantage returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var dataResp avResponse
	if err := json.NewDecoder(resp.Body).Decode(&dataResp); err != nil {
		return nil, fmt.Errorf("failed to decode alpha vantage schema: %w", err)
	}

	if dataResp.TimeSeries == nil {
		return nil, fmt.Errorf("alpha vantage payload missing 'Time Series (1min)'. Might be rate limited")
	}

	var timestamps []string
	for ts := range dataResp.TimeSeries {
		timestamps = append(timestamps, ts)
	}

	// AV returns newest first. We need oldest first for backtest
	sort.Strings(timestamps)

	var bars []Bar
	for _, ts := range timestamps {
		tickData := dataResp.TimeSeries[ts]

		parsedTS, err := ParseTimestamp(ts)
		if err != nil {
			// AV specific: it lacks timezone offset in string, default UTC or EST
			// Since our ParseTimestamp falls back to 2006-01-02, a raw 2006-01-02 15:04:00 needs custom handling
			// We augment it here if it's strictly space-separated
			if len(ts) == 19 { // YYYY-MM-DD HH:MM:SS
				parsedTS, err = time.Parse("2006-01-02 15:04:05", ts)
			}
			if err != nil {
				log.Printf("Invalid AV timestamp %q: %v", ts, err)
				continue
			}
		}

		open, _ := ParseDecimal(tickData["1. open"])
		high, _ := ParseDecimal(tickData["2. high"])
		low, _ := ParseDecimal(tickData["3. low"])
		closeVal, _ := ParseDecimal(tickData["4. close"])
		volume, _ := ParseDecimal(tickData["5. volume"])

		if open == 0 || closeVal == 0 {
			continue // skip empty
		}

		bars = append(bars, Bar{
			Timestamp: parsedTS,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeVal,
			Volume:    volume,
		})
	}

	return bars, nil
}

func (h *AlphaVantageDataHandler) Stream(visitor func(b Bar, rowIdx int) bool) error {
	reqUrl := h.EndpointURL
	if reqUrl == "" {
		reqUrl = fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=%s&interval=1min&outputsize=compact&apikey=%s", h.Symbol, h.APIKey)
	}

	var lastSeen time.Time
	rowIdx := 0

	log.Printf("Bootstrapping Historical Data via Alpha Vantage REST for %s...", h.Symbol)
	
	// Phase 1: Historical Setup
	bars, err := h.fetchAndParse(reqUrl)
	if err != nil {
		return fmt.Errorf("initial historical bootstrap failed: %w", err)
	}

	for _, b := range bars {
		if cont := visitor(b, rowIdx); !cont {
			return nil
		}
		lastSeen = b.Timestamp
		rowIdx++
	}

	log.Printf("Historical buffer flushed (%d bars). Entering live polling mode...", rowIdx)

	// Phase 2: Live Polling loop
	pollInterval := h.PollInterval
	if pollInterval == 0 {
		pollInterval = 60 * time.Second // Default 1 min poll spacing to respect rate limits
	}

	polls := 0
	for {
		if h.MaxPollCount > 0 && polls >= h.MaxPollCount {
			break // test terminator
		}
		polls++

		time.Sleep(pollInterval)

		// Fetch standard payload
		newBars, err := h.fetchAndParse(reqUrl)
		if err != nil {
			log.Printf("Alpha Vantage Polling skipped due to error: %v", err)
			continue
		}

		for _, b := range newBars {
			if b.Timestamp.After(lastSeen) {
				if cont := visitor(b, rowIdx); !cont {
					return nil
				}
				lastSeen = b.Timestamp
				rowIdx++
			}
		}
	}

	return nil
}

// Load methods omitted/unsupported for direct live handlers
func (h *AlphaVantageDataHandler) Load() ([]Bar, error) {
	return nil, fmt.Errorf("Load not supported for AlphaVantage stream mode")
}

func (h *AlphaVantageDataHandler) LoadHead(n int) ([]Bar, error) {
	return nil, fmt.Errorf("LoadHead not supported for AlphaVantage stream mode")
}

func (h *AlphaVantageDataHandler) LoadTail(n int) ([]Bar, error) {
	return nil, fmt.Errorf("LoadTail not supported for AlphaVantage stream mode")
}

func (h *AlphaVantageDataHandler) LoadRange(start, end int) ([]Bar, error) {
	return nil, fmt.Errorf("LoadRange not supported for AlphaVantage stream mode")
}

func (h *AlphaVantageDataHandler) LoadStats() (Stats, error) {
	return Stats{}, fmt.Errorf("LoadStats not supported for AlphaVantage stream mode")
}
