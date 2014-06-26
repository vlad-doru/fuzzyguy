"""This app has the sole purpose of providing a user friendly interface
for the user to interact with the fuzzyguy service which is in fact
a command-line webserver designed for fast fuzzy queries"""

import traceback
import sys
import requests
import json

from flask import Flask
from flask import request
from test import test_service

app = Flask(__name__, static_url_path='/static')

# We make this kind of global varaible in order to speed up the
# communication between the app and the server
session = requests.Session()


@app.route("/fuzzy/cleardemo")
def cleardemo():
    url = "http://localhost:8080/fuzzy"
    params = {
        "store": "demostore"
    }
    r = session.delete(url, params=params)
    r = session.post(url, params=params)
    return "Successfully deleted"


@app.route("/fuzzy/loadenglish")
def load_english():
    with open("data/english.dat") as f:
        url = "http://localhost:8080/fuzzy"
        params = {
            "store": "demostore"
        }
        r = session.delete(url, params=params)
        r = session.post(url, params)

        if r.status_code != 201:
            response.status_code = r.status_code
            return "Error on creating the store"

        batch = {}

        for line in f:
            word, defintion = line.split('\t')
            batch[word] = defintion
            if len(batch) == 1000:
                try:
                    r = session.put("http://localhost:8080/fuzzy/batch", params={
                        "store": "demostore",
                        "dictionary": json.dumps(batch)
                    })
                    if r.status_code != 200:
                        response.status_code = r.status_code
                        return "Error on putting the key-value batch"
                except Exception, e:
                    print e
                batch = {}

    return "Successfully loaded the dictionary"


def query_service(distance):
    url = "http://localhost:8080/fuzzy"
    params = {
        "store": "demostore",
        "distance": distance,
        "results": "5",
        "key": request.args.get('key'),
    }
    r = session.get(url, params=params)
    if r.status_code != 200:
        response.status_code = r.status_code
        return "Error while quering the service"
    return r.text


@app.route("/fuzzy/query")
def query():
    return query_service(2)


@app.route("/fuzzy/exact")
def exact():
    return query_service(0)


@app.route("/fuzzy/add", methods=['PUT'])
def add():
    url = "http://localhost:8080/fuzzy"
    try:
        r = session.put(url, params=request.form)
    except Exception, e:
        print e

    return r.text


@app.route("/")
def index():
    return app.send_static_file('index.html')


@app.route("/fuzzy/test")
def test():
    try:
        size = request.args["testsize"]
        results = request.args["results"]
        distance = request.args["distance"]

        return test_service(size, distance, results)

    except Exception, e:
        print e
        print traceback.format_exc()


@app.route("/monitor")
def monitor():
    return app.send_static_file('monitor.html')

@app.route("/profiler")
def profiler():
    return app.send_static_file('profiler.html')
