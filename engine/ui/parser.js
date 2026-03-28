function parseLogs(csvString) {
    const lines = csvString.trim().split('\n');
    const bars = [];
    const signals = [];

    // Skip header
    for (let i = 1; i < lines.length; i++) {
        const fields = lines[i].split(',');
        if (fields.length < 2) continue;

        const timestamp = fields[0];
        const eventType = fields[1];
        
        // Convert to 'YYYY-MM-DD' for lightweight charts
        const time = timestamp.split('T')[0];

        if (eventType === 'BAR') {
            const inds = {};
            if (fields[10]) {
                const parts = fields.slice(10).join(',').split('|');
                parts.forEach(p => {
                    const kv = p.split(':');
                    if (kv.length === 2 && kv[0]) {
                        inds[kv[0]] = parseFloat(kv[1]);
                    }
                });
            }

            bars.push({
                time: time,
                open: parseFloat(fields[5]),
                high: parseFloat(fields[6]),
                low: parseFloat(fields[7]),
                close: parseFloat(fields[8]),
                indicators: inds
            });
        } else if (eventType === 'BUY' || eventType === 'SELL') {
            const price = parseFloat(fields[2]);
            const isBuy = eventType === 'BUY';
            signals.push({
                time: time,
                position: isBuy ? 'belowBar' : 'aboveBar',
                color: isBuy ? '#26a69a' : '#ef5350',
                shape: isBuy ? 'arrowUp' : 'arrowDown',
                text: eventType,
                price: price
            });
        }
    }

    return { bars, signals };
}

if (typeof module !== 'undefined') {
    module.exports = { parseLogs };
}
