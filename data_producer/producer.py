import asyncio
import csv
import json
import logging
from pathlib import Path
import websockets

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger("DataProducer")

CSV_PATH = Path(__file__).parent.parent / "AAPL.csv"
DELAY = 0.05  # sending one row every 50ms

async def stream_data(websocket):
    logger.info(f"Client connected: {websocket.remote_address}")
    try:
        with open(CSV_PATH, 'r') as f:
            reader = csv.DictReader(f)
            count = 0
            for row in reader:
                # Format payload to roughly match what the Go engine expects
                payload = {
                    "timestamp": f"{row['Date']}T00:00:00Z", # convert YYYY-MM-DD to RFC3339 daily
                    "open": row['Open'],
                    "high": row['High'],
                    "low": row['Low'],
                    "close": row['Close'],
                    "volume": row['Volume'],
                }
                await websocket.send(json.dumps(payload))
                count += 1
                await asyncio.sleep(DELAY)
        logger.info(f"Finished streaming {count} rows")
    except websockets.exceptions.ConnectionClosed:
        logger.info(f"Client disconnected: {websocket.remote_address}")
    except Exception as e:
        logger.error(f"Error streaming data: {e}")

async def main():
    if not CSV_PATH.exists():
        logger.error(f"CSV file not found at {CSV_PATH.resolve()}")
        return

    logger.info("Starting WebSocket server on ws://localhost:8080")
    async with websockets.serve(stream_data, "localhost", 8080):
        await asyncio.Future()  # run forever

if __name__ == "__main__":
    asyncio.run(main())
