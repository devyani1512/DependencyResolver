from flask import Flask
import numpy as np
import psycopg2  

app = Flask(__name__)

@app.route('/')
def hello():
    return "Flask + Numpy + PostgreSQL!"

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=8000)
