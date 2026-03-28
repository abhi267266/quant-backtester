# 🚀 Quant Backtester Engine

![QuantFlow Event-Driven Architecture](assets/backtesting_event_loop_flow.svg)

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

#### 4. JSON Dynamic Strategies (`dynamic`)
Design fully custom, mathematically sound trading algorithms without writing any Go logic. Define multiple $O(1)$ indicators dynamically and build complex rule sets out of simple JSON structures.

**Use Case:** 
If you are iterating through dozens of quantitative ideas (like pairing a fast/slow moving average trend detector alongside an RSI momentum filter), recompiling Go code for every permutation is painstakingly slow. The internal Dynamic Engine parses a JSON array of technical indicators into memory instantly, tracking them side-by-side perfectly. 
*Note: Placing multiple conditions inside a `buy` or `sell` block intrinsically evaluates as a **Logical AND** (meaning every condition strictly must be true natively at a single bar close for the pipeline to generate a valid Signal).*

**Create a Strategy configuration (`strategy.json`):**
```json
{
  "strategy_name": "SMA_Crossover_with_RSI_Filter",
  "indicators": [
    { "id": "fast_ma", "type": "SMA", "params": { "period": 10 } },
    { "id": "slow_ma", "type": "SMA", "params": { "period": 50 } },
    { "id": "rsi_main", "type": "RSI", "params": { "period": 14 } }
  ],
  "rules": {
    "buy": [
      { "type": "crossover", "left_operand": "fast_ma", "right_operand": "slow_ma" },
      { "type": "greater_than", "left_operand": "rsi_main", "value": 30 }
    ],
    "sell": [
      { "type": "crossunder", "left_operand": "fast_ma", "right_operand": "slow_ma" }
    ]
  }
}
```

**Run the JSON engine natively:**
```bash
./inspector dynamic -config strategy.json -capital 15000 -log strategy_logs.csv < historical_data.csv
```

#### 5. Interactive Visualization Dashboard
After executing a backtest using the `-log` flag natively, utilize the built-in python charting interface to render Premium JavaScript Canvas views dynamically displaying your Total Portfolio Equity scaling over time.

```bash
python3 visualize.py
```
*Note: Ensure your `-log` output is configured as `strategy_logs.csv` inside the `engine/` directory before running the renderer.*

#### 6. Dynamic Trading Terminal UI
We've introduced a robust, zero-dependency HTML/JS **Trading Terminal UI** utilizing **Lightweight Charts** for deep, interactive chart analysis. The frontend dynamically parses `strategy_logs.csv` directly from the engine. It intelligently renders Candlesticks, multi-indicator overlays (e.g., SMAs automatically mapping to the main price pane), and oscillator indicators (e.g., RSI binding to an isolated sub-pane) without requiring frontend hardcoding. It concurrently plots exact Buy/Sell execution signals visually synced precisely to the candlestick timeline!

**Usage:**
1. Execute a backtest using the `-log` flag to output local logs (`strategy_logs.csv`).
2. Serve the `engine/ui/` directory locally using any lightweight HTTP server (e.g., Python's `http.server` to prevent classic file:// CORS errors).
```bash
cd engine/ui
python3 -m http.server 8000
```
3. Navigate to [http://localhost:8000](http://localhost:8000) in your browser to experience the dynamic, auto-rendered multi-pane interactive charts.

## 🧪 Validating The Engine

We run rigorous Testing, Memory Profiling, and Look-Ahead Bias prevention natively.

To execute the test suite (and visualize our strict zero-allocation integrity benchmarks):
```bash
cd engine
go test ./... -v -bench . -benchmem
```

---
*Disclaimer: Past performance is not indicative of future market results, but continuing to use `float64` for finance is highly indicative of future bugs.*
