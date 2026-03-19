# 🚀 Quant Backtester Engine (Phase 1)

Welcome to the **Quant Backtester Engine**! If you're here because floating-point precision errors destroyed your last algorithm and cost you a small yacht, you've come to the right place. 🛥️📉

We don't do `float64` here. We do **hardcore, scaled `int64` fixed-point math** ($10^8$ precision—down to the saturating satoshi!). Why? Because `0.1 + 0.2` shouldn't equal `0.30000000000000004` when real money is on the line. 

## 🛠 What's Inside?

Currently, this repository houses **Phase 1** of our high-performance trading engine architecture:
- **`data` Package:** A lightning-fast, zero-allocation-per-bar CSV stream reader. It handles huge datasets like a champ without storing everything in RAM.
- **`indicators` Package:** SMA, EMA, and RSI built directly on top of 128-bit integer math and Wilder's True Sum. This guarantees your moving averages never drift.
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

## 🧪 Testing and Proving Zero Allocations

We love TDD almost as much as we love integer division. Run the tests to see those sweet, sweet zero-allocation benchmarks proving our internal processing loop runs in the low microseconds:

```bash
cd engine
go test ./... -v -bench . -benchmem
```

---
*Disclaimer: Past performance is not indicative of future market results, but continuing to use `float64` for finance is highly indicative of future bugs.*
