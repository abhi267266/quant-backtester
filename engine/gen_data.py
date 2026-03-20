import csv
import math
from datetime import datetime, timedelta

def generate_dummy_data(filename="historical_data.csv", rows=1000):
    start_time = datetime(2024, 1, 1)
    base_price = 100.0
    
    with open(filename, mode='w', newline='') as file:
        writer = csv.writer(file)
        # Header matches your CSVDataHandler expectations
        writer.writerow(["Timestamp", "Open", "High", "Low", "Close", "Volume"])
        
        for i in range(rows):
            # Create a sine wave + some noise for "trending" behavior
            # This ensures we get crossovers for your SMA strategy
            trend = math.sin(i / 50.0) * 20 
            noise = (i % 10) * 0.5
            
            current_close = base_price + trend + noise
            current_open = current_close - (noise / 2)
            current_high = max(current_open, current_close) + 1.5
            current_low = min(current_open, current_close) - 1.5
            current_volume = 1000 + (i * 2)
            
            timestamp = (start_time + timedelta(days=i)).strftime("%Y-%m-%dT%H:%M:%SZ")
            
            writer.writerow([
                timestamp,
                f"{current_open:.8f}",
                f"{current_high:.8f}",
                f"{current_low:.8f}",
                f"{current_close:.8f}",
                f"{current_volume:.8f}"
            ])

if __name__ == "__main__":
    generate_dummy_data()
    print("Generated historical_data.csv with 1000 rows.")