package logger

import (
	"bufio"
	"io"
	"time"
)

// LogWriter defines the interface for logging trades and snapshots without allocating memory
type LogWriter interface {
	LogTrade(t TradeEntry) error
	LogSnapshot(s SnapshotEntry) error
	Flush() error
}

type TradeEntry struct {
	Timestamp  time.Time
	Side       string
	Price      int64
	Qty        int64
	TotalValue int64
}

type SnapshotEntry struct {
	Timestamp     time.Time
	TotalEquity   int64
	Cash          int64
	UnrealizedPnL int64
}

// WriteFixedPoint writes an int64 scaled by 10^8 to a bufio.Writer directly to avoid allocations.
func WriteFixedPoint(w *bufio.Writer, val int64) error {
	if val == 0 {
		_, err := w.WriteString("0.00000000")
		return err
	}

	uval := uint64(val)
	if val < 0 {
		if err := w.WriteByte('-'); err != nil {
			return err
		}
		uval = uint64(-val)
	}

	intPart := uval / 100000000
	fracPart := uval % 100000000

	if intPart == 0 {
		if err := w.WriteByte('0'); err != nil {
			return err
		}
	} else {
		var buf [24]byte
		pos := 24
		for intPart > 0 {
			pos--
			buf[pos] = '0' + byte(intPart%10)
			intPart /= 10
		}
		if _, err := w.Write(buf[pos:]); err != nil {
			return err
		}
	}

	if err := w.WriteByte('.'); err != nil {
		return err
	}

	var fbuf [8]byte
	for i := 7; i >= 0; i-- {
		fbuf[i] = '0' + byte(fracPart%10)
		fracPart /= 10
	}
	_, err := w.Write(fbuf[:])
	return err
}

// CSVLogger implements LogWriter out to an io.Writer using bufio exactly avoiding memory allocations
type CSVLogger struct {
	w       *bufio.Writer
	timeBuf [32]byte
}

func NewCSVLogger(w io.Writer) *CSVLogger {
	bw := bufio.NewWriter(w)
	// Write the CSV header describing both trade and snapshot alignments
	bw.WriteString("Timestamp,EventType,Price_or_Equity,Qty_or_Cash,TotalValue_or_UnrealizedPnL\n")
	return &CSVLogger{w: bw}
}

// WriteTime manually formats the time to RFC3339 via byte appends cleanly
func (c *CSVLogger) WriteTime(t time.Time) error {
	res := t.AppendFormat(c.timeBuf[:0], time.RFC3339)
	_, err := c.w.Write(res)
	return err
}

func (c *CSVLogger) LogTrade(t TradeEntry) error {
	if err := c.WriteTime(t.Timestamp); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if _, err := c.w.WriteString(t.Side); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if err := WriteFixedPoint(c.w, t.Price); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}

	// Qty is in units, multiplying by 10^8 to standardize it for WriteFixedPoint output format
	if err := WriteFixedPoint(c.w, t.Qty*100000000); err != nil {
		return err
	}

	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if err := WriteFixedPoint(c.w, t.TotalValue); err != nil {
		return err
	}
	if err := c.w.WriteByte('\n'); err != nil {
		return err
	}
	return nil
}

func (c *CSVLogger) LogSnapshot(s SnapshotEntry) error {
	if err := c.WriteTime(s.Timestamp); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if _, err := c.w.WriteString("SNAPSHOT"); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if err := WriteFixedPoint(c.w, s.TotalEquity); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if err := WriteFixedPoint(c.w, s.Cash); err != nil {
		return err
	}
	if err := c.w.WriteByte(','); err != nil {
		return err
	}
	if err := WriteFixedPoint(c.w, s.UnrealizedPnL); err != nil {
		return err
	}
	if err := c.w.WriteByte('\n'); err != nil {
		return err
	}
	return nil
}

func (c *CSVLogger) Flush() error {
	return c.w.Flush()
}

// NoOpLogger mocks the LogWriter for pure speed benchmark execution
type NoOpLogger struct{}

func (n *NoOpLogger) LogTrade(t TradeEntry) error       { return nil }
func (n *NoOpLogger) LogSnapshot(s SnapshotEntry) error { return nil }
func (n *NoOpLogger) Flush() error                      { return nil }
