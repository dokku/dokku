"""Independent WebSocket HTTP-upgrade check, used after the browser verdict."""
import base64
import hashlib
import json
import os
import socket
import struct


def websocket():
  connection = socket.create_connection(('localhost', 19080), timeout=15)
  stream = connection.makefile('rb')
  def headers():
    first = stream.readline().decode().strip()
    fields = {}
    while True:
      line = stream.readline().decode().strip()
      if not line: break
      key, value = line.split(':', 1)
      fields[key.lower()] = value.strip()
    return first, fields
  try:
    key = base64.b64encode(os.urandom(16)).decode()
    connection.sendall(('GET /ws HTTP/1.1\r\nHost: localhost:19080\r\nUpgrade: websocket\r\n'
      'Connection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: ' + key + '\r\n\r\n').encode())
    first, fields = headers()
    expected = base64.b64encode(hashlib.sha1((key + '258EAFA5-E914-47DA-95CA-C5AB0DC85B11').encode()).digest()).decode()
    if ' 101 ' not in first or fields.get('sec-websocket-accept') != expected:
      raise RuntimeError('Upgrade: ' + first)
    message = ('transport-probe-' + os.urandom(8).hex()).encode()
    mask = os.urandom(4)
    connection.sendall(bytes([0x81, 0x80 | len(message)]) + mask + bytes(b ^ mask[i % 4] for i, b in enumerate(message)))
    start = stream.read(2)
    length = start[1] & 127
    if length == 126: length = struct.unpack('!H', stream.read(2))[0]
    data = json.loads(stream.read(length))
    if data['echo'] != message.decode(): raise RuntimeError('Echo mismatch')
    return {'upgrade': first, 'echoMatched': True, 'upstreamHost': data['host'], 'marker': data['marker']}
  finally:
    stream.close()
    connection.close()
