#!/usr/bin/env python
# coding:utf-8

import sys
import CallbackServer

def websockets(query):
    str = 'hello'
    s = str.encode('utf-8')
    return s

if __name__ == '__main__':
    PORT = 50000
    CallbackServer.start(PORT, websockets)
    serverURL = 'http://127.0.0.1'
    gameID = '8d25c551-d275-4fb5-948e-2baa48f32a7a'
    boardURL = 'http://localhost:3000/?engine='+serverURL+':'+PORT+'&game='+gameID+'&autoplay=true'
