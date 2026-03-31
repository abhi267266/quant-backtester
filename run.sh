#!/bin/bash

# Navigate into the engine directory where the inspector binary natively executes
cd engine || exit 1

# Execute 12 years of deep historical bounds fetching locally from Yahoo Finance cleanly bypassing limits
./inspector backtest \
  -mode yfinance \
  -symbol AAPL \
  -start "2014-01-01" \
  -end "2026-03-31" \
  -log strategy_logs.csv \
  -config strategy.json
