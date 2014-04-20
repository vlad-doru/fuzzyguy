import requests
import timeit
from progressbar import *


def readData(filename):
    with open(filename) as f:
        content = [line.strip() for line in f.readlines()]
    nr_keys = int(content[0])
    keys = content[1: 1 + nr_keys]
    queries = [query.split('\t')
               for query in content[2 + nr_keys:]]
    return keys, queries


def main():
    time = 0
    keys, queries = readData("../fuzzy/data/testset_5000.dat")

    req_params = {
        'store': 'fuzzytest',
        'key': 'cheie',
        'value': 'testval',
    }
    url = 'http://localhost:8080/fuzzy'
    s = requests.Session()

    # Initializaing store
    r = s.post(url, req_params)

    widgets = [
        'Putting keys: ', Percentage(), ' ', Bar(marker=RotatingMarker()),
        ' ', ETA(), ' ', FileTransferSpeed()]
    pbar = ProgressBar(widgets=widgets, maxval=len(keys)).start()

    for i, key in enumerate(keys):
        req_params['key'] = key
        try:
            r = s.put(url, req_params)
            print(r)
            time += r.elapsed.microseconds
        except Exception as e:
            print(e)
        pbar.update(i)
    pbar.finish()

    print(
        "Average time for a put request is {0} miliseconds".format(time / (len(keys) * 1000)))

if __name__ == '__main__':
    main()
