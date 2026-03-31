package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockAVServer(t *testing.T, payloads []string) *httptest.Server {
	callIdx := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if callIdx < len(payloads) {
			fmt.Fprintln(w, payloads[callIdx])
			callIdx++
		} else {
			// fallback if they keep requesting
			fmt.Fprintln(w, payloads[len(payloads)-1])
		}
	}))
}

func TestAlphaVantageDataHandler_Stream(t *testing.T) {
	// First request simulates historical backup
	payload1 := `{
    "Meta Data": {
        "1. Information": "Intraday (1min) open, high, low, close prices and volume",
        "2. Symbol": "IBM",
        "3. Last Refreshed": "2023-11-24 19:59:00",
        "4. Interval": "1min",
        "5. Output Size": "Compact",
        "6. Time Zone": "US/Eastern"
    },
    "Time Series (1min)": {
        "2023-11-24 19:59:00": {
            "1. open": "155.0000",
            "2. high": "155.5000",
            "3. low": "154.5000",
            "4. close": "154.8000",
            "5. volume": "100"
        },
        "2023-11-24 19:58:00": {
            "1. open": "154.9000",
            "2. high": "154.9500",
            "3. low": "154.8500",
            "4. close": "154.9100",
            "5. volume": "200"
        }
    }
}`

	// Second payload simulates the first continuous poll picking up a new minute tick
	payload2 := `{
    "Meta Data": {
        "1. Information": "Intraday (1min) open, high, low, close prices and volume"
    },
    "Time Series (1min)": {
        "2023-11-24 20:00:00": {
            "1. open": "154.8000",
            "2. high": "154.9000",
            "3. low": "154.7000",
            "4. close": "154.8500",
            "5. volume": "500"
        },
        "2023-11-24 19:59:00": {
            "1. open": "155.0000",
            "2. high": "155.5000",
            "3. low": "154.5000",
            "4. close": "154.8000",
            "5. volume": "100"
        }
    }
}`

	server := mockAVServer(t, []string{payload1, payload2})
	defer server.Close()

	handler := &AlphaVantageDataHandler{
		Symbol:       "IBM",
		APIKey:       "demo",
		EndpointURL:  server.URL,
		PollInterval: 5 * time.Millisecond,
		MaxPollCount: 1, // Restrict to one polling cycle for test to prevent infinite loop
	}

	var bars []Bar
	err := handler.Stream(func(b Bar, rowIdx int) bool {
		bars = append(bars, b)
		return true
	})

	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	// We expect 3 unique bars: 19:58, 19:59, 20:00
	if len(bars) != 3 {
		t.Fatalf("Expected 3 bars, got %d", len(bars))
	}

	// Bars should be sorted ascending by timestamp
	// 19:58:00
	if bars[0].Open != 15490000000 {
		t.Errorf("First bar open mismatch: %v", bars[0].Open)
	}
	// 19:59:00
	if bars[1].Open != 15500000000 {
		t.Errorf("Second bar open mismatch: %v", bars[1].Open)
	}
	// 20:00:00
	if bars[2].Open != 15480000000 {
		t.Errorf("Third bar open mismatch: %v", bars[2].Open)
	}
}

func TestAlphaVantageDataHandler_UnsupportedMethods(t *testing.T) {
	handler := &AlphaVantageDataHandler{Symbol: "IBM"}

	if _, err := handler.LoadHead(10); err == nil {
		t.Error("Expected error for LoadHead")
	}
	if _, err := handler.LoadTail(10); err == nil {
		t.Error("Expected error for LoadTail")
	}
}
