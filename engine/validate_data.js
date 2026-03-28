const fs = require('fs');
const { parseLogs } = require('./ui/parser.js');
const data = fs.readFileSync('strategy_logs.csv', 'utf8');
const { bars, signals } = parseLogs(data);

console.log("Checking bars for Lightweight Charts constraints...");
let lastTime = '';
for (let i = 0; i < bars.length; i++) {
    const b = bars[i];
    if (Object.values(b).some(v => Number.isNaN(v))) {
        console.error("NaN found at bar", i, b);
    }
    if (lastTime && b.time <= lastTime) {
        console.error("Bars not strictly ascending! Bar", i, b.time, "<=", lastTime);
    }
    lastTime = b.time;
}

console.log("Checking signals...");
let lastSigTime = '';
for (let i = 0; i < signals.length; i++) {
    const s = signals[i];
    if (Number.isNaN(s.price)) {
         console.error("NaN found at signal", i, s);
    }
    if (lastSigTime && s.time < lastSigTime) {
         console.error("Signals not ascending!", s.time, "<", lastSigTime);
    }
    lastSigTime = s.time;
}
console.log("Validation complete.");
