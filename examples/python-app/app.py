from flask import Flask
import psycopg


app = Flask(__name__)

@app.route('/')
def hello():
    return "This is your dependency resolver!"

if __name__ == '__main__':
    app.run(debug=True)