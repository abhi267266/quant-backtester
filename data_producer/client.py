import asyncio
import websockets
import json

async def listen():
    uri = "ws://localhost:8080"
    try:
        async with websockets.connect(uri) as ws:
            print(f"Successfully connected to {uri}\n")
            
            # Receive and print the first 3 lines
            for i in range(10):
                message = await ws.recv()
                data = json.loads(message)
                print(f"Row {i+1}:")
                # Pretty print the incoming JSON
                print(json.dumps(data, indent=2))
                print("-" * 20)
                
    except ConnectionRefusedError:
        print(f"Failed to connect to {uri}. Is the server running?")

if __name__ == "__main__":
    asyncio.run(listen())
