package strategy

import (
	"github.com/quant-backtester/engine/data"
)

type Action string

const (
    Buy  Action = "BUY"
    Sell Action = "SELL"
    Hold Action = "HOLD"
)

type Signal struct {
    Action Action
    Price  int64 // The price at which the signal was generated
}

type Strategy interface {
	OnBar(bar data.Bar) Signal
}



