document.addEventListener('DOMContentLoaded', () => {
    const fileInput = document.getElementById('csvFileInput');
    const loadBtn = document.getElementById('loadBtn');
    
    // Chart options configuring the TradingView Dark Mode Theme
    const chartOptions = {
        layout: {
            textColor: '#d1d4dc',
            background: { type: 'solid', color: '#131722' },
        },
        grid: {
            vertLines: { color: 'rgba(42, 46, 57, 0.5)' },
            horzLines: { color: 'rgba(42, 46, 57, 0.5)' },
        },
        crosshair: {
            mode: LightweightCharts.CrosshairMode.Normal,
        },
        rightPriceScale: {
            borderColor: 'rgba(197, 203, 206, 0.8)',
            autoScale: true,
        },
        timeScale: {
            borderColor: 'rgba(197, 203, 206, 0.8)',
            rightOffset: 12,
            barSpacing: 10,
        },
    };

    // Initialize Main Chart (OHLC + SMA)
    const mainChartContainer = document.getElementById('mainChart');
    const mainChart = LightweightCharts.createChart(mainChartContainer, chartOptions);
    mainChart.resize(mainChartContainer.clientWidth, mainChartContainer.clientHeight);
    
    const candleSeries = mainChart.addCandlestickSeries({
        upColor: '#26a69a',
        downColor: '#ef5350',
        borderVisible: false,
        wickUpColor: '#26a69a',
        wickDownColor: '#ef5350',
    });

    // Initialize Indicator Chart (RSI)
    const rsiChartContainer = document.getElementById('rsiChart');
    const rsiChart = LightweightCharts.createChart(rsiChartContainer, {
        ...chartOptions,
        rightPriceScale: {
            ...chartOptions.rightPriceScale,
            autoScale: true,
            scaleMargins: { top: 0.1, bottom: 0.1 },
        }
    });
    rsiChart.resize(rsiChartContainer.clientWidth, rsiChartContainer.clientHeight);

    // Synchronize horizontal scrolling/zooming between the two panes
    mainChart.timeScale().subscribeVisibleTimeRangeChange((range) => {
        if (range !== null && range.from !== null && range.to !== null) {
            try {
                rsiChart.timeScale().setVisibleRange(range);
            } catch (e) {}
        }
    });
    rsiChart.timeScale().subscribeVisibleTimeRangeChange((range) => {
        if (range !== null && range.from !== null && range.to !== null) {
            try {
                mainChart.timeScale().setVisibleRange(range);
            } catch (e) {}
        }
    });

    // Handle Window Resize
    window.addEventListener('resize', () => {
        mainChart.resize(mainChartContainer.clientWidth, mainChartContainer.clientHeight);
        rsiChart.resize(rsiChartContainer.clientWidth, rsiChartContainer.clientHeight);
    });

    const indicatorSeriesMap = {};

    function renderData(bars, signals) {
        if (bars.length > 0) {
            candleSeries.setData(bars);
            
            const allIndicators = new Set();
            bars.forEach(b => {
                if (b.indicators) Object.keys(b.indicators).forEach(k => allIndicators.add(k));
            });

            const colors = ['#2962ff', '#e91e63', '#f57c00', '#9c27b0', '#00bcd4', '#009688', '#4caf50', '#cddc39'];
            let colorIdx = 0;

            allIndicators.forEach(indName => {
                const indData = bars.map(b => ({ time: b.time, value: b.indicators[indName] }))
                                   .filter(d => d.value !== undefined && !isNaN(d.value) && d.value > 0);

                if (!indicatorSeriesMap[indName]) {
                    const isOscillator = indName.toLowerCase().includes('rsi') || indName.toLowerCase().includes('osc');
                    const targetChart = isOscillator ? rsiChart : mainChart;
                    
                    const series = targetChart.addLineSeries({
                        color: colors[colorIdx++ % colors.length],
                        lineWidth: 2,
                        crosshairMarkerVisible: false,
                        title: indName,
                    });

                    if (isOscillator) {
                        series.createPriceLine({ price: 70, color: '#ef5350', lineWidth: 1, lineStyle: LightweightCharts.LineStyle.Dashed, axisLabelVisible: false });
                        series.createPriceLine({ price: 30, color: '#26a69a', lineWidth: 1, lineStyle: LightweightCharts.LineStyle.Dashed, axisLabelVisible: false });
                    }
                    indicatorSeriesMap[indName] = series;
                }

                indicatorSeriesMap[indName].setData(indData);
            });
        }

        if (signals.length > 0) {
            candleSeries.setMarkers(signals);
        }
        // Disabled fitContent() on every render so the user can freely zoom/pan while live streaming.
        // mainChart.timeScale().fitContent(); 
    }

    function fetchData() {
        const cacheBuster = new Date().getTime();
        fetch(`../strategy_logs.csv?cb=${cacheBuster}`)
            .then(res => res.text())
            .then(csvData => {
                const { bars, signals } = parseLogs(csvData);
                renderData(bars, signals);
            })
            .catch(err => console.error("Failed to load logs:", err));
    }

    // Fetch immediately and then auto-poll every 5 seconds for live mode
    fetchData();
    setInterval(fetchData, 5000);

    // Setup Load Data file input (Optional fallback)
    loadBtn.addEventListener('click', () => {
        if (!fileInput.files.length) {
            alert('Please select a strategy_logs.csv file first');
            return;
        }

        const file = fileInput.files[0];
        const reader = new FileReader();

        reader.onload = (e) => {
            const csvData = e.target.result;
            const { bars, signals } = parseLogs(csvData);
            renderData(bars, signals);
        };
        reader.readAsText(file);
    });
});
