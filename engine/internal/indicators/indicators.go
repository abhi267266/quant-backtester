package indicators

import (
	"errors"
	"github.com/quant-backtester/engine/data"
)

const Scale int64 = 100000000

var ErrInsufficientData = errors.New("insufficient data for indicator period")

type BatchIndicator interface {
	Compute(bars []data.Bar) ([]int64, error)
}

type StatefulIndicator interface {
	Update(bar data.Bar) (int64, error)
}

