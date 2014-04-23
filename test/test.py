import requests
import timeit
from progressbar import *
import json


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


def main():
    time = 0
    keys, queries = readData("../fuzzy/data/testset_100000.dat")

    req_params = {
        'store': 'fuzzytest',
    }
    url = 'http://localhost:8080/fuzzy'
    s = requests.Session()

    # Initializaing store
    r = s.post(url, req_params)
    url = 'http://localhost:8080/fuzzy/batch'

    widgets = [
        'Putting keys: ', Percentage(), ' ', Bar(marker=RotatingMarker()),
        ' ', ETA(), ' ', FileTransferSpeed()]
    batch_size = 10000
    pbar = ProgressBar(widgets=widgets, maxval=len(keys) / batch_size).start()

    i = 0

    for l in chunks(keys, batch_size):
        dic = {}
        for key in l:
            dic[key] = "test"
        req_params["dictionary"] = json.dumps(dic)

        r = s.put(url, req_params)
        i += 1
        time += r.elapsed.microseconds

        if r.status_code != 200:
            print(r.status_code, r.text)
            break

        pbar.update(i)
    pbar.finish()

    print(
        "Average time for a batch put request of {0} key-value pairs is {1} miliseconds".format(batch_size, time / (i * 1000)))

    url = 'http://localhost:8080/fuzzy'

    widgets = [
        'Putting keys: ', Percentage(), ' ', Bar(marker=RotatingMarker()),
        ' ', ETA(), ' ', FileTransferSpeed()]
    pbar = ProgressBar(widgets=widgets, maxval=len(queries)).start()

    req_params = {
        'store': 'fuzzytest',
        'distance': 3,
        'results': 5
    }

    time = 0

    for i, query in enumerate(queries):
        correct, queried = query
        req_params['key'] = queried
        r = s.get(url, params=req_params)
        if r.status_code != 200:
            print(r, r.text)
            break
        time += r.elapsed.microseconds
        pbar.update(i)
    pbar.finish()
    print("Average time for a get request with {0} keys datastore is {1} miliseconds".format(
        len(keys), time / (len(queries) * 1000)))


if __name__ == '__main__':
    main()
