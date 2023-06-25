#!/usr/bin/env python

import requests
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse

def start(port, callback):
    def handler(*args):
        CallbackServer(callback, *args)
    server = HTTPServer(('', int(port)), handler)
    server.serve_forever()

class CallbackServer(BaseHTTPRequestHandler):
    def __init__(self, callback, *args):
        self.callback = callback
        BaseHTTPRequestHandler.__init__(self, *args)

    def do_GET(self):
        parsed_path = urlparse(self.path)
        query = parsed_path.query
        self.send_response(200)
        self.end_headers()
        result = self.callback(query)
        message = '\r\n'.join(result)
        self.wfile.write(message)
        return
