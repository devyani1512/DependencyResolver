from flask import Flask
#  import psycopg2 - this line also triggers docker service
# testing if this comment detects PostgreSQL
app = Flask(__name__)

@app.route('/')
def hello():
    return "This is your dependency resolver!"

if __name__ == '__main__':
    app.run(debug=True)