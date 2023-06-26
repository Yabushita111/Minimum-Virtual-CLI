
import threading
from http.server import HTTPServer, SimpleHTTPRequestHandler
import asyncio
import websockets
import webbrowser


filePATH = '/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/'
gameID = '8d25c551-d275-4fb5-948e-2baa48f32a7a'
filepath = filePATH + gameID + ".json"

async def send(websocket):
    with open(filepath) as f:
        async for s_line in f:
            await websocket.send(s_line)

async def main():
    async with websockets.serve(send, '127.0.0.1',50000):
        server_thread2 = threading.Thread(target= asyncio.Future())
        server_thread2.start()

class Handler(SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        asyncio.run(main())
        return

def begin_dummy_server():
    hostIp = "127.0.0.1"
    port = 0
    httpd = HTTPServer((hostIp, port), Handler)
    server_thread = threading.Thread(target=httpd.serve_forever)
    server_thread.start()

if __name__ == "__main__":
    begin_dummy_server()
    PORT = '50000'
    serverURL = 'http://127.0.0.1'
    boardURL = 'http://localhost:3000/?engine='+serverURL+':'+PORT+'&game='+gameID+'&autoplay=true'
    webbrowser.open(boardURL)
    
    