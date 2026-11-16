import json
import os
import socket
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

ROLE = sys.argv[1]
MARKER = os.environ.get('E2E_MARKER', 'dokku-backend-unconfigured')
HOSTNAME = socket.gethostname()
EXTERNAL_PREFIX = os.environ.get('E2E_EXTERNAL_PREFIX', '').rstrip('/')
PAGE = b'''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Dokku routing probe</title>
<script src="/probe.js" defer></script></head>
<body><h1>Dokku routing probe</h1><button id="check">Check backend</button>
<p id="result" role="status">Backend not checked</p>
<label for="message">Message</label><input id="message">
<button id="echo">Send WebSocket message</button><p id="ws-result">WebSocket not checked</p>
<p id="http-host"></p>
<a href="/api/v0/identity">Open routed page</a>
<a href="/api/v0/redirect">Open API login</a></body></html>'''
SCRIPT = b'''document.querySelector('#check').addEventListener('click', async () => {
  const result = document.querySelector('#result');
  result.textContent = 'Checking backend';
  try {
    const response = await fetch('/api/v0/identity', {cache: 'no-store'});
    if (!response.ok) throw new Error('Backend HTTP ' + response.status);
    const data = await response.json();
    if (data.role !== 'api') throw new Error('Wrong upstream');
    result.textContent = data.marker + ' Container: ' + data.hostname;
  } catch (error) {
    result.textContent = 'Backend check failed: ' + error.message;
  }
});
document.querySelector('#echo').addEventListener('click', () => {
  const result = document.querySelector('#ws-result');
  result.textContent = 'Connecting';
  const socket = new WebSocket('ws://' + location.host + '/ws');
  socket.onopen = () => socket.send(document.querySelector('#message').value);
  socket.onmessage = event => {
    const data = JSON.parse(event.data);
    result.textContent = 'Echo: ' + data.echo + ' Marker: ' + data.marker + ' WebSocket Host: ' + data.host;
    socket.close();
  };
  socket.onerror = () => { result.textContent = 'WebSocket failed'; };
});
fetch('/probe-host').then(r => r.json()).then(data => {
  document.querySelector('#http-host').textContent = 'HTTP Host: ' + data.host;
});'''


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if ROLE == 'web' and self.path == '/':
            payload, content_type, status = PAGE, 'text/html; charset=utf-8', 200
        elif ROLE == 'web' and self.path == '/probe.js':
            payload, content_type, status = SCRIPT, 'text/javascript; charset=utf-8', 200
        elif self.path == '/probe-host':
            payload = json.dumps({'role': ROLE, 'host': self.headers.get('Host')}).encode()
            content_type, status = 'application/json', 200
        elif ROLE == 'web' and self.path == '/healthz':
            payload, content_type, status = b'web ready', 'text/plain', 200
        elif ROLE == 'api' and self.path in ('/api/v0/redirect', '/redirect'):
            self.send_response(302)
            self.send_header('Location', EXTERNAL_PREFIX + '/login')
            self.send_header('Content-Length', '0')
            self.end_headers()
            return
        elif ROLE == 'api' and self.path == '/login':
            payload, content_type, status = b'<h1>API login page</h1>', 'text/html', 200
        elif ROLE == 'api' and self.path == '/':
            payload, content_type, status = b'<h1>API root page</h1>', 'text/html', 200
        elif ROLE == 'api' and self.path in ('/api/v0/identity', '/identity'):
            payload = json.dumps({'role': ROLE, 'marker': MARKER, 'hostname': HOSTNAME}).encode()
            content_type, status = 'application/json', 200
        else:
            payload, content_type, status = b'Wrong upstream or path', 'text/plain', 404
        self.send_response(status)
        self.send_header('Content-Type', content_type)
        self.send_header('Cache-Control', 'no-store')
        self.send_header('Content-Length', str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)


ThreadingHTTPServer(('0.0.0.0', 5000 if ROLE == 'web' else 5001), Handler).serve_forever()
