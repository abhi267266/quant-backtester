package data

import (
	"strings"
	"testing"
	"time"
)

func TestCSVDataHandler_Load(t *testing.T) {
	csvData := `Timestamp,Open,High,Low,Close,Volume
2023-01-01T15:04:05Z,100.0,105.0,99.0,104.5,1000
2023-01-02,104.5,106.0,103.0,105.0,2000`

	reader := strings.NewReader(csvData)
	handler := &CSVDataHandler{Reader: reader}

	bars, err := handler.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}

	// Validate Bar 1
	expectedTime1, _ := time.Parse(time.RFC3339, "2023-01-01T15:04:05Z")
	b1 := bars[0]
	if b1.Timestamp != expectedTime1 {
		t.Errorf("expected time %v, got %v", expectedTime1, b1.Timestamp)
	}
	if b1.Open != 10000000000 || b1.High != 10500000000 || b1.Low != 9900000000 || b1.Close != 10450000000 || b1.Volume != 100000000000 {
		t.Errorf("expected varying OHLCV values for bar 1, got %+v", b1)
	}

	// Validate Bar 2
	expectedTime2, _ := time.Parse("2006-01-02", "2023-01-02")
	b2 := bars[1]
	if b2.Timestamp != expectedTime2 {
		t.Errorf("expected time %v, got %v", expectedTime2, b2.Timestamp)
	}
	if b2.Open != 10450000000 || b2.High != 10600000000 || b2.Low != 10300000000 || b2.Close != 10500000000 || b2.Volume != 200000000000 {
		t.Errorf("expected varying OHLCV values for bar 2, got %+v", b2)
	}
}
