import json
from flask import Flask, jsonify, request
import requests
import os

app = Flask(__name__)

@app.route('/')
def index():
    return jsonify({'active': True})

@app.route('/healthz/')
def healthz():
    return jsonify({'healthy': True})

@app.route('/hello/')
def hello():
    name = request.args.get('name', 'World')
    return f"Hello {name}!"

@app.route('/proxy/', methods=['POST'])
def proxy():
    request_data = request.get_json()
    
    host = request_data.get('host', 'http://postman-echo.com').strip('/')
    args = request_data.get('args', 'get?foo1=bar1&foo2=bar2').strip('/')
    r = requests.get(f"{host}/{args}")

    return r.text

if __name__ == "__main__":
    port = int(os.environ.get('PORT', 9090))
    app.run(debug=True, host='0.0.0.0', port=port)
