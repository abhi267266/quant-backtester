# 🚀 Quant Backtester Engine

Welcome to the **Quant Backtester Engine**! This is a high-performance, strictly zero-allocation, fixed-point integer trading engine written in Go. If you're here because floating-point precision errors disrupted your last algorithm, you've come to the right place. 🛥️📉

We don't do `float64` for critical math. We use scaled `int64` fixed-point representations ($10^8$ precision—down to the saturating satoshi!). This prevents classical rounding inaccuracies like `0.1 + 0.2 = 0.30000000000000004` when real money is on the line.

## 🛠 Internal Architecture

The pipeline is completely engineered on an **Event-Driven Architecture (EDA)** to eliminate Garbage Collector (GC) pressure by operating in an asynchronous, decoupled, $O(1)$ streaming fashion:

- **`event` Package (The Core EDA):** The central nervous system of the backtester. An ultra-fast, channel-based `EventBus` that handles data decoupling. It standardizes asynchronous communication matching live trading mechanics using core event variants: `MarketEvent`, `SignalEvent`, `OrderEvent`, and `FillEvent`.
- **`data` Package:** A lightning-fast, zero-allocation-per-bar CSV stream reader. It evaluates massive datasets iteratively without storing the entire file into RAM, publishing `MarketEvent`s directly to the bus.
- **`indicators` Package:** Features technical indicators (SMA, EMA, RSI) built natively on $10^8$ integer math. It leverages a Stateful `Update()` pipeline that evaluates indicators incrementally slice-free per tick—producing 1,400x speedups natively over batch processing.
- **`portfolio` Package:** A completely zero-allocation, $O(1)$ portfolio accounting layer tracking fixed-point Cash, Cost Basis, Peak Equity, Realized PnL, Max Drawdown, and precise Position Sizes continuously across trades. It is equipped with strict positional guards against "ghost signals". It listens to `SignalEvent`s and `FillEvent`s.
- **`logger` Package:** Streams $10^8$ fixed-point values completely separated from the heap memory using a manual fixed-point digit extraction format into CSV files, circumventing Go's alloc-heavy `fmt.Sprintf` completely. Contains a Benchmark-safety `NoOpLogger` validating 0 allocs/op bounds.
- **`strategy` Package:** Sandbox environment housing your algorithms (e.g., `SMACrossover`). Strategies listen to the event bus for `MarketEvent`s, process price action completely devoid of look-ahead bias, and publish `SignalEvent`s back.

## 🕵️‍♂️ Using the CLI Inspector

Because staring at raw gigabyte CSV dumps is difficult, we built a terminal CLI executable (`inspector`). It streams data intuitively, generates formatted ASCII tables natively, logs live trade data predictably, and reconstructs strict internal integers cleanly into human-readable prices for debugging.

### Building the Executable

Navigate to the `engine` directory and compile the binary natively:

```bash
cd engine
go build -o inspector main.go
```

The data handler expects a standard CSV format feeding via `stdin`: `Timestamp, Open, High, Low, Close, Volume`.

### Available Commands

#### 1. Data Inspection (`inspect`)
Interact directly with the raw CSV streams:
- **Head/Tail Extraction:**
  ```bash
  ./inspector inspect head -n 10 < historical_data.csv
  ./inspector inspect tail -n 10 < historical_data.csv
  ```
- **Range Extraction (Surgical evaluation):**
  ```bash
  ./inspector inspect range -start 50 -end 65 < historical_data.csv
  ```
- **Stats & Continuity Checks:** Automatically scans for missing intervals and internal timestamp gaps without blowing up RAM.
  ```bash
  ./inspector inspect stats < historical_data.csv
  ```

#### 2. Technical Indicators (`sma`, `ema`, `rsi`)
Compute math-heavy technicals directly across the tail bounds of the dataset:
```bash
./inspector sma -period 14 -n 10 < historical_data.csv
./inspector ema -period 14 -n 10 < historical_data.csv
./inspector rsi -period 14 -n 10 < historical_data.csv
```

#### 3. Full Strategy Backtesting (`backtest`)
Run a robust `SMACrossover` strategy synchronously through the zero-allocation `portfolio` handler. This outputs a beautiful real-time execution table tracking `Timestamp`, `Action`, `Price`, `Capital`, `Cash`, `CostBasis`, and `PositionSize` for every executed signal, flawlessly followed by a definitive Performance Summary metric table.

**Configurable Flags:**
- `-short`: Fast SMA period (default: `5`)
- `-long`: Slow SMA period (default: `10`)
- `-capital`: Initial starting simulated capital (default: `10000.0`)
- `-log`: Write detailed raw snapshot and trade outputs into a specific CSV filename
- `-interval`: How frequently (in bars) to log absolute snapshot frames

**Example usage:**
```bash
./inspector backtest -short 5 -long 20 -capital 25000 -interval 50 -log backtest_logs.csv < historical_data.csv
```

## 🧪 Validating The Engine

We run rigorous Testing, Memory Profiling, and Look-Ahead Bias prevention natively.

To execute the test suite (and visualize our strict zero-allocation integrity benchmarks):
```bash
cd engine
go test ./... -v -bench . -benchmem
```

---
*Disclaimer: Past performance is not indicative of future market results, but continuing to use `float64` for finance is highly indicative of future bugs.*
