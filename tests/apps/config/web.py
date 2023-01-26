import http.server
import json
import os
import sys


class GetHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.end_headers()

        if self.path == '/':
            self.wfile.write("python/http.server".encode("utf-8"))
        elif self.path == '/conftest':
            value = os.getenv("CONFTEST", "")
            self.wfile.write(value.encode("utf-8"))
        elif self.path == '/hello':
            value = os.getenv("HELLO", "")
            self.wfile.write(value.encode("utf-8"))
        else:
            data = json.dumps(dict(os.environ), sort_keys=True, indent=4)
            self.wfile.write(data.encode("utf-8"))


if __name__ == "__main__":
    if os.getenv("FAIL_ON_STARTUP") == "true":
        print("Failing on startup due to FAIL_ON_STARTUP=true")
        sys.exit(1)

    for arg in sys.argv:
        print(f"Arg: {arg}")

    port = int(os.getenv("PORT", 5000))
    server = http.server.HTTPServer(("0.0.0.0", port), GetHandler)
    print("Listening on port {0}".format(port))
    server.serve_forever()
