import os
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
        return 'python/flask'
