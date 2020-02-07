import os
from flask import Flask, render_template
try:
    from werkzeug.wsgi import SharedDataMiddleware
except ImportError:
    from werkzeug.middleware.shared_data import SharedDataMiddleware

app = Flask(__name__)

@app.route('/')
def hello_world():
    return render_template('index.html')

app.wsgi_app = SharedDataMiddleware(app.wsgi_app, { '/': os.path.join(os.path.dirname(__file__), 'static') })
app.wsgi_app = SharedDataMiddleware(app.wsgi_app, { '/': os.path.join(os.path.dirname(__file__), 'static/.tmp') })

if __name__ == '__main__':
    app.run(host='0.0.0.0', debug=True)
