"""This APP has the sole purpose of providing a user friendly interface
for the user to interact with the fuzzyguy service which is in fact
a command-line webserver designed for fast fuzzy queries"""

import traceback
import requests
import json

from flask import Flask
from flask import request
from test import test_service

APP = Flask(__name__, static_url_path='/static')

# We make this kind of global varaible in order to speed up the
# communication between the APP and the server
SESSION = requests.Session()


@APP.route("/fuzzy/cleardemo")
def cleardemo():
    """ We remove the demostore from the service """
    url = "http://localhost:8080/fuzzy"
    params = {
        "store": "demostore"
    }
    SESSION.delete(url, params=params)
    SESSION.post(url, params=params)
    return "Successfully deleted"


@APP.route("/fuzzy/loadenglish")
def load_english():
    """ We load the english dictionary and place it in the demo store """
    with open("data/english.dat") as eng_file:
        url = "http://localhost:8080/fuzzy"
        params = {
            "store": "demostore"
        }
        res = SESSION.delete(url, params=params)
        res = SESSION.post(url, params)

        if res.status_code != 201:
            return "Error on creating the store", res.status_code

        batch = {}

        for line in eng_file:
            word, defintion = line.split('\t')
            batch[word] = defintion
            if len(batch) == 1000:
                try:
                    res = SESSION.put("http://localhost:8080/fuzzy/batch",
                                    params={
                                        "store": "demostore",
                                        "dictionary": json.dumps(batch)
                                    })
                    if res.status_code != 200:
                        return "Error on pute key-value batch", res.status_code
                except Exception:
                    pass
                batch = {}

    return "Successfully loaded the dictionary"


def query_service(distance):
    """ We query the service through the API """
    url = "http://localhost:8080/fuzzy"
    params = {
        "store": "demostore",
        "distance": distance,
        "results": "5",
        "key": request.args.get('key'),
    }
    res = SESSION.get(url, params=params)
    if res.status_code != 200:
        return "Error while quering the service", res.status_code
    return res.text


@APP.route("/fuzzy/query")
def query():
    """ We query at a distance of 2 """
    return query_service(2)


@APP.route("/fuzzy/exact")
def exact():
    """ We make exact queries """
    return query_service(0)


@APP.route("/fuzzy/add", methods=['PUT'])
def add():
    """ We add a key-value pair """
    url = "http://localhost:8080/fuzzy"
    try:
        res = SESSION.put(url, params=request.form)
    except Exception:
        pass

    return res.text


@APP.route("/")
def index():
    """ We send the demo main page """
    return APP.send_static_file('index.html')


@APP.route("/fuzzy/test")
def test():
    """ We benchmark the service """
    try:
        size = request.args["testsize"]
        results = request.args["results"]
        distance = request.args["distance"]

        return test_service(size, distance, results)

    except Exception:
        print traceback.format_exc()


@APP.route("/monitor")
def monitor():
    """ We return the statis file corresponding for testing """
    return APP.send_static_file('monitor.html')


@APP.route("/profiler")
def profiler():
    """ We serve the profiler html """
    return APP.send_static_file('profiler.html')
