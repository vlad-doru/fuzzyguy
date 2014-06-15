from __future__ import print_function
from __future__ import division

import requests
import timeit
import json
import sys


def readData(filename):
    with open(filename) as f:
        content = [line.strip() for line in f.readlines()]
    nr_keys = int(content[0])
    keys = content[1: 1 + nr_keys]
    queries = [query.split('\t')
               for query in content[2 + nr_keys:]]
    return keys, queries


def chunks(l, n):
    """ Yield successive n-sized chunks from l."""
    for i in range(0, len(l), n):
        yield l[i:i + n]


def test_service(size, distance, results):
    time = 0
    keys, queries = readData(
        "data/testset_{0}.dat".format(size))

    req_params = {
        'store': 'fuzzytest',
    }
    url = 'http://localhost:8080/fuzzy'
    s = requests.Session()

    # Initializaing store
    r = s.post(url, req_params)
    url = 'http://localhost:8080/fuzzy/batch'

    batch_size = 100000

    i = 0

    for l in chunks(keys, batch_size):
        dic = {}
        for key in l:
            dic[key] = "test"
        req_params["dictionary"] = json.dumps(dic)

        r = s.put(url, req_params)
        i += 1
        time += (r.elapsed.seconds * 1000000) + r.elapsed.microseconds

        if r.status_code != 200:
            break

    batch_time = time / (len(keys) / batch_size)
    batch_total = time

    req_params = {
        'store': 'fuzzytest',
        'distance': distance,
        'results': results,
        'keys': json.dumps([query[1] for query in queries])
    }

    r = s.get(url, params=req_params)
    micro = r.elapsed.microseconds / 1000
    seconds = r.elapsed.seconds
    time = 1000 * seconds + micro

    result = json.loads(r.text)
    accuracy = 0
    for c, r in zip([query[0] for query in queries], result):
        try:
            index = r.index(c)
            accuracy += (len(r) - index) / len(r)
        except Exception as e:
            accuracy += 0

    stats = {
        'time': round(time),
        'keys': len(keys),
        'queries': len(queries),
        'accuracy': round((accuracy * 100) / len(queries), 2),
        'batch_size': batch_size,
        'batch_time': round(batch_time / 1000),
        'batch_total': round(batch_total / 1000),
        'distance': distance,
        'results': results,
        'throughput': round((1000 / (time)) * len(queries))
    }

    # We must delete the database
    url = 'http://localhost:8080/fuzzy'
    req_params = {
        'store': 'fuzzytest'
    }
    r = s.delete(url, params=req_params)
    return json.dumps(stats)