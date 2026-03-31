package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// YFinanceDataHandler pulls OHLCV data directly from Yahoo Finance JSON charts
type YFinanceDataHandler struct {
	Symbol    string
	StartDate time.Time
	EndDate   time.Time

	// Internal dependencies for mockability
	EndpointURL string
}

// ConvertFloatToScaled safely transforms Yahoo's raw float64 values into fixed-point $10^8 format
func ConvertFloatToScaled(val float64) int64 {
	if val < 0 {
		return int64(val*float64(Decimals) - 0.5)
	}
	return int64(val*float64(Decimals) + 0.5)
}

func (h *YFinanceDataHandler) Load() ([]Bar, error) { return nil, fmt.Errorf("Load not natively supported") }
func (h *YFinanceDataHandler) LoadHead(n int) ([]Bar, error) { return nil, fmt.Errorf("LoadHead not supported") }
func (h *YFinanceDataHandler) LoadTail(n int) ([]Bar, error) { return nil, fmt.Errorf("LoadTail not supported") }
func (h *YFinanceDataHandler) LoadRange(s, e int) ([]Bar, error) { return nil, fmt.Errorf("LoadRange not supported") }
func (h *YFinanceDataHandler) LoadStats() (Stats, error) { return Stats{}, fmt.Errorf("LoadStats not supported") }

type yfResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

func (h *YFinanceDataHandler) Stream(visitor func(b Bar, rowIdx int) bool) error {
	baseURL := os.Getenv("YAHOO_FINANCE_URL")
	if baseURL == "" {
		baseURL = "https://query2.finance.yahoo.com/v8/finance/chart/"
	}

	reqURL := fmt.Sprintf("%s%s?period1=%d&period2=%d&interval=1d", baseURL, h.Symbol, h.StartDate.Unix(), h.EndDate.Unix())
	if h.EndpointURL != "" {
		reqURL = h.EndpointURL // Overridden natively for offline structured testing bounds
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// Simulate organic browser cleanly to bypass HTTP 401 Unauthorized blockages
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("yahoo finance http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	var data yfResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("yahoo finance json decode error: %w", err)
	}

	if data.Chart.Error != nil {
		return fmt.Errorf("yahoo finance api error: %v", data.Chart.Error)
	}

	if len(data.Chart.Result) == 0 {
		return fmt.Errorf("yahoo finance returned entirely empty result gracefully securely")
	}

	result := data.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return fmt.Errorf("missing indicators structurally securely inside mapping")
	}

	quote := result.Indicators.Quote[0]
	length := len(result.Timestamp)

	for i := 0; i < length; i++ {
		// YFinance sometimes internally skips slices mapping cleanly as `null` throwing indices out of bound gracefully.
		// Protect explicitly to prevent crashing
		if i >= len(quote.Open) || i >= len(quote.High) || i >= len(quote.Low) || i >= len(quote.Close) || i >= len(quote.Volume) {
			continue 
		}

		b := Bar{
			Timestamp: time.Unix(result.Timestamp[i], 0).UTC(),
			Open:      ConvertFloatToScaled(quote.Open[i]),
			High:      ConvertFloatToScaled(quote.High[i]),
			Low:       ConvertFloatToScaled(quote.Low[i]),
			Close:     ConvertFloatToScaled(quote.Close[i]),
			Volume:    ConvertFloatToScaled(quote.Volume[i]),
		}

		if !visitor(b, i) {
			break
		}
	}

	return nil
}
