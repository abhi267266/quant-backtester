package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockYFinanceServer(t *testing.T, payload string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, payload)
	}))
}

func TestConvertFloatToScaled(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected int64
	}{
		{"Zero", 0.0, 0},
		{"Basic Positive", 150.25, 15025000000},
		{"Basic Negative", -150.25, -15025000000},
		{"High Precision", 123.45678912, 12345678912},
		{"Rounding Up", 123.456789126, 12345678913}, // Should round up the 8th decimal
		{"Rounding Down", 123.456789124, 12345678912},
		{"Smallest Positive", 0.00000001, 1},
		{"Smallest Negative", -0.00000001, -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertFloatToScaled(tc.input)
			if result != tc.expected {
				t.Errorf("ConvertFloatToScaled(%f) = %d; expected %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestYFinanceDataHandler_Stream(t *testing.T) {
	payload := `{
  "chart": {
    "result": [
      {
        "timestamp": [1704205800, 1704292200],
        "indicators": {
          "quote": [
            {
              "open": [187.14999389648438, 184.22000122070312],
              "high": [188.44000244140625, 185.8800048828125],
              "low": [183.88999938964844, 183.42999267578125],
              "close": [185.63999938964844, 184.25],
              "volume": [82488700, 58414500]
            }
          ]
        }
      }
    ],
    "error": null
  }
}`

	server := mockYFinanceServer(t, payload)
	defer server.Close()

	handler := &YFinanceDataHandler{
		Symbol:      "AAPL",
		EndpointURL: server.URL,
	}

	var bars []Bar
	err := handler.Stream(func(b Bar, rowIdx int) bool {
		bars = append(bars, b)
		return true
	})

	if err != nil {
		t.Fatalf("Stream failed natively: %v", err)
	}

	if len(bars) != 2 {
		t.Fatalf("Expected strictly 2 completely parsed bars, got %d", len(bars))
	}

	expectedDay1Open := ConvertFloatToScaled(187.14999389648438)
	if bars[0].Open != expectedDay1Open {
		t.Errorf("Mismatch on Day 1 Open conversion: expected %d, got %d", expectedDay1Open, bars[0].Open)
	}
	
	expectedDay2Volume := ConvertFloatToScaled(float64(58414500))
	if bars[1].Volume != expectedDay2Volume {
		t.Errorf("Mismatch on Day 2 Volume natively: expected %d, got %d", expectedDay2Volume, bars[1].Volume)
	}
}
