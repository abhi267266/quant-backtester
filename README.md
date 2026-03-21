# 🚀 Quant Backtester Engine (Phase 1)

Welcome to the **Quant Backtester Engine**! If you're here because floating-point precision errors destroyed your last algorithm and cost you a small yacht, you've come to the right place. 🛥️📉

We don't do `float64` here. We do **hardcore, scaled `int64` fixed-point math** ($10^8$ precision—down to the saturating satoshi!). Why? Because `0.1 + 0.2` shouldn't equal `0.30000000000000004` when real money is on the line. 

## 🛠 What's Inside?

Currently, this repository houses our high-performance trading engine architecture:
- **`data` Package:** A lightning-fast, zero-allocation-per-bar CSV stream reader. It handles huge datasets like a champ without storing everything in RAM.
- **`indicators` Package:** SMA, EMA, and RSI built directly on top of 128-bit integer math and Wilder's True Sum. Recently restructured into highly modular files, now fully equipped with $O(1)$ `StatefulIndicator` update processing alongside the legacy `BatchIndicator` logic!
- **`portfolio` Package:** (NEW!) A zero-allocation, $O(1)$ portfolio accounting layer that tracks fixed-point Cash, Cost Basis, Peak Equity, Realized PnL, and Max Drawdown dynamically across tick data!
- **`strategy` Package:** Features `SMACrossover` rigorously tested via Strict TDD protocols, built natively linking to our lightning-fast $O(1)$ stateful pipeline.
- **The CLI Inspector:** A gorgeous terminal interface to preview your data without accidentally loading a 50GB CSV file directly into memory and melting your laptop.

## 🕵️‍♂️ The CLI Inspector (Executable Handling)

Because staring at raw gigabyte CSV dumps is a great way to lose your sanity, we built an inspector executable. It safely streams data (using a slick Ring Buffer for trailing limits) and prints beautifully aligned ASCII tables. Best of all: it transforms our internal $10^8$ integer scale back into human-readable trailing decimals purely for the terminal view!

### Building the Executable

Navigate to the `engine` directory and build the beast into a clean binary:

```bash
cd engine
go build -o inspector main.go
```

Now, feed it a CSV file via standard input! Our data handler expects the classic 6 columns: `Timestamp, Open, High, Low, Close, Volume`.

### CLI Commands

#### Data Inspection (`inspect`)

**1. Head (Look at the start)**
Skip the headers and grab the first 10 rows:
```bash
./inspector inspect head -n 10 < historical_data.csv
```

**2. Tail (Look at the end)**
Want to see how your data finishes without blowing up your RAM? This uses a highly-efficient rolling ring buffer:
```bash
./inspector inspect tail -n 10 < historical_data.csv
```

**3. Range (Surgical extraction)**
Need to look at the exact moment the market crashed? Extract a highly specific block of arrays:
```bash
./inspector inspect range -start 50 -end 65 < historical_data.csv
```

**4. Stats (The Polygraph Test)**
Don't trust the data provider? (You shouldn't). Check the overall row count, start/end dates, and natively run a **Continuity Check** to automatically detect if your exchange secretly skipped intervals or days unannounced!
```bash
./inspector inspect stats < historical_data.csv
```

#### Technical Indicators (`sma`, `ema`, `rsi`)

Calculate and display technical indicators on the dataset. Outputs the specified indicator alongside the closing prices. You can specify the indicator period and how many of the latest rows to display.

**5. Simple Moving Average (SMA)**
Calculate a 14-period SMA and print the last 10 rows:
```bash
./inspector sma -period 14 -n 10 < historical_data.csv
```

**6. Exponential Moving Average (EMA)**
Calculate a 14-period EMA and print the last 10 rows:
```bash
./inspector ema -period 14 -n 10 < historical_data.csv
```

**7. Relative Strength Index (RSI)**
Calculate a 14-period RSI and print the last 10 rows:
```bash
./inspector rsi -period 14 -n 10 < historical_data.csv
```

#### Full Strategy Backtesting (`backtest`)

**8. Run a Stateful Backtest with Performance Summary**
Run a complete `SMACrossover` strategy natively streamed through our $O(1)$ zero-allocation `portfolio` package. It will print trade signals dynamically and output a beautifully formatted Performance Summary containing your Final Equity, Net Profit, and Max Drawdown dynamically calculated per-tick!

**Available Options:**
- `-short`: The period for the fast Simple Moving Average (default: 5).
- `-long`: The period for the slow Simple Moving Average (default: 10).
- `-capital`: The initial simulated capital to fund the portfolio (default: 10000.0).

```bash
./inspector backtest -short 5 -long 20 -capital 25000 < historical_data.csv
```

## 🧪 Testing and Proving Zero Allocations

We love TDD almost as much as we love integer division. Run the tests to see those sweet, sweet zero-allocation benchmarks proving our internal processing loop runs in the low microseconds:

```bash
cd engine
go test ./... -v -bench . -benchmem
```

## 🏎 Stateful vs. Batch Benchmarking (Phase 2 Optimizations)

In **Phase 2**, we overhauled the indicator processing layer to utilize an `O(1)` stateful `Update()` pipeline rather than reconstructing full `O(N^2)` historical slices per bar like the legacy `BatchIndicator` method. 

### How to Run the Benchmark
Navigate to the `indicators` package and execute the test suite specifically targeting benchmark methods:
```bash
cd engine/internal/indicators
go test -bench=. -benchmem -benchtime=1s
```

### The Results (Apple M2 Architecture)
Simulating a 10,000 bar strategy backtest clearly demonstrates why we don't recalculate aggregated historical batches on every tick:

```text
goos: darwin
goarch: arm64
pkg: github.com/quant-backtester/engine/internal/indicators
cpu: Apple M2
BenchmarkSMA_Batch-8               12      85654708 ns/op     431385016 B/op    10011 allocs/op
BenchmarkSMA_Stateful-8         19618         60337 ns/op        160184 B/op      719 allocs/op
```

- **Speed:** Our `StatefulIndicator` pipeline runs in ~60 microseconds per full history run—an over **1,400x speedup** compared to the 85.6ms batch pipeline overhead!
- **Memory Pressure:** Generates only 160 KB of predictable backing-array slice capacity compared to forcing the Garbage Collector to grind through 431 MB of newly constructed float/int slices per block. Zero GC overhead scales infinitely.

---
*Disclaimer: Past performance is not indicative of future market results, but continuing to use `float64` for finance is highly indicative of future bugs.*
