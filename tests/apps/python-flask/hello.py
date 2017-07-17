import os
from flask import Flask, abort

app = Flask(__name__)

@app.route('/')
def hello():
        return 'python/flask'


@app.route('/<env>')
def get_enc(env):
        val = os.environ.get(env, None)
        if val:
                return val
        else:
                abort(404)
