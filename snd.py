#!/usr/bin/env python

import time
import socket
import random


TCP_IP = 'localhost'
TCP_PORT = 7777
BUFFER_SIZE = 10  # Normally 1024, but we want fast response

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.bind((TCP_IP, TCP_PORT))
s.listen(1)

conn, addr = s.accept()
print('Connection address:', addr)
while 1:
    data = conn.recv(BUFFER_SIZE)
    if not data: break
    time.sleep(1)
    print("received data:", data)
    conn.send(data)  # echo
conn.close()
