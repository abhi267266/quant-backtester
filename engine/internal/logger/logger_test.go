package logger

import (
	"bufio"
	"bytes"
	"io"
	"testing"
	"time"
)

func TestCSVStreaming(t *testing.T) {
	var buf bytes.Buffer
	l := NewCSVLogger(&buf)

	now := time.Now().UTC()
	err := l.LogTrade(TradeEntry{
		Timestamp:  now,
		Side:       "BUY",
		Price:      300000000,
		Qty:        5,
		TotalValue: 1500000000,
	})
	if err != nil {
		t.Fatal(err)
	}
	l.Flush()

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("BUY")) {
		t.Errorf("expected output to contain BUY, got %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("3.00000000")) {
		t.Errorf("expected output to contain 3.00000000, got %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("5.00000000")) {
		t.Errorf("expected output to contain QTY 5.00000000, got %s", output)
	}
}

func TestFixedPointFormatting(t *testing.T) {
	cases := []struct {
		val      int64
		expected string
	}{
		{0, "0.00000000"},
		{123456789, "1.23456789"},
		{100000000, "1.00000000"},
		{5000000, "0.05000000"},
		{-123456789, "-1.23456789"},
	}

	for _, c := range cases {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		err := WriteFixedPoint(w, c.val)
		if err != nil {
			t.Fatal(err)
		}
		w.Flush()
		if b.String() != c.expected {
			t.Errorf("expected %s, got %s", c.expected, b.String())
		}
	}
}

func TestSnapshotTrigger(t *testing.T) {
	var buf bytes.Buffer
	l := NewCSVLogger(&buf)

	interval := 3
	for i := 0; i < 10; i++ {
		if i%interval == 0 {
			l.LogSnapshot(SnapshotEntry{
				TotalEquity: int64(i * 100000000),
			})
		}
	}
	l.Flush()

	if bytes.Count(buf.Bytes(), []byte("SNAPSHOT")) != 4 {
		t.Errorf("expected exactly 4 snapshots, got %d", bytes.Count(buf.Bytes(), []byte("SNAPSHOT")))
	}
}

func BenchmarkCSVLogger(b *testing.B) {
	l := NewCSVLogger(io.Discard)
	now := time.Now()
	entry := TradeEntry{
		Timestamp:  now,
		Side:       "BUY",
		Price:      123450000,
		Qty:        10,
		TotalValue: 1234500000,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.LogTrade(entry)
	}
}

func BenchmarkNoOpLogger(b *testing.B) {
	var l LogWriter = &NoOpLogger{}
	now := time.Now()
	entry := TradeEntry{
		Timestamp:  now,
		Side:       "BUY",
		Price:      123450000,
		Qty:        10,
		TotalValue: 1234500000,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.LogTrade(entry)
	}
}
