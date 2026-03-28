const { parseLogs } = require('./parser');

describe('Log Parser', () => {
    test('parses BAR events correctly', () => {
        const csvData = `Timestamp,EventType,Price_or_Equity,Qty_or_Cash,TotalValue_or_UnrealizedPnL,Open,High,Low,Close,Volume,IndicatorsMetadata
2024-01-01T00:00:00Z,BAR,,,,100.0,101.5,98.5,100.0,1000.0,fast_ma:105.5|slow_ma:100.0`;
        const result = parseLogs(csvData);
        
        expect(result.bars.length).toBe(1);
        expect(result.bars[0]).toEqual({
            time: '2024-01-01',
            open: 100.0,
            high: 101.5,
            low: 98.5,
            close: 100.0,
            indicators: {
                'fast_ma': 105.5,
                'slow_ma': 100.0
            }
        });
    });

    test('parses BUY and SELL signals correctly', () => {
        const csvData = `Timestamp,EventType,Price_or_Equity,Qty_or_Cash,TotalValue_or_UnrealizedPnL,Open,High,Low,Close,Volume,SMA9,RSI14
2024-01-16T00:00:00Z,BUY,108.41,92.00,9973.75,,,,,,,
2024-01-18T00:00:00Z,SELL,112.50,92.00,10200.0,,,,,,,`;
        const result = parseLogs(csvData);
        
        expect(result.signals.length).toBe(2);
        expect(result.signals[0]).toEqual({
            time: '2024-01-16',
            position: 'belowBar',
            color: '#26a69a',
            shape: 'arrowUp',
            text: 'BUY',
            price: 108.41
        });
        expect(result.signals[1]).toEqual({
            time: '2024-01-18',
            position: 'aboveBar',
            color: '#ef5350',
            shape: 'arrowDown',
            text: 'SELL',
            price: 112.50
        });
    });

    test('ignores SNAPSHOT events', () => {
        const csvData = `Timestamp,EventType,Price_or_Equity,Qty_or_Cash,TotalValue_or_UnrealizedPnL,Open,High,Low,Close,Volume,SMA9,RSI14
2024-01-01T00:00:00Z,SNAPSHOT,10000.0,10000.0,0.0,,,,,,,`;
        const result = parseLogs(csvData);
        
        expect(result.bars.length).toBe(0);
        expect(result.signals.length).toBe(0);
    });
});
