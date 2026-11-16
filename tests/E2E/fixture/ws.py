import json
import os
from websockets.sync.server import serve


def echo(connection):
  for message in connection:
    connection.send(json.dumps({'echo': message, 'marker': os.environ['E2E_MARKER'],
                               'host': connection.request.headers.get('Host')}))


with serve(echo, '0.0.0.0', 5002) as server:
  server.serve_forever()
