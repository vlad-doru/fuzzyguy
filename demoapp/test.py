""" We define here the functions we need to test the performance of our service.
    We interact with the server via its RESTful API by using requests module """

from __future__ import print_function
from __future__ import division

import json
import requests
import uuid


def read_data(filename):
    """ Open and read information from the testfile with the given name

    Args:
        filename (string): the name of the test file we want to read

    Returns:
        (list[string], list[string]) whcih represent the keys which should be
        placed in the service while the second list represent the test queries
        we should execute in order to evaluate the performance of the service.
    """
    with open(filename) as test_file:
        content = [line.strip() for line in test_file.readlines()]
    nr_keys = int(content[0])
    keys = content[1: 1 + nr_keys]
    queries = [query.split('\t')
               for query in content[2 + nr_keys:]]
    return keys, queries


def chunks(list_object, size):
    """ Yield successive n-sized chunks from l.

        Args:
            list_object (list): list to be sharded
            size (int): maximum number of items in a shard

        Yields:
            list: lists from list_object with the maximum specified file until
                  the inital list has been consumed."""
    for index in range(0, len(list_object), size):
        yield list_object[index:index + size]


def create_store(test_session, store_name):
    """ We create a database only for benchmarking the service """
    test_session.post('http://localhost:8080/fuzzy',
                      {
                          'store': store_name,
                      })


def delete_store(test_session, store_name):
    """ We delete the database we previously created for benchmarking """
    test_session.delete('http://localhost:8080/fuzzy',
                        params={
                            'store': store_name
                        })


def upload_keys(test_session, store_name, keys, stats):
    """ We upload the initial keys to the database

    Args:
        test_session (session): the session we use to interact with the service
        keys (list): the keys which we should upload
        store_name (string): the name of the store we want to interact with
        stats (dict): a dictionary where we collect statistics about testing """

    url = 'http://localhost:8080/fuzzy/batch'
    req_params = {
        'store': store_name,
    }
    batch_size = 100000

    time = 0
    for shard in chunks(keys, batch_size):
        dic = {}
        for key in shard:
            dic[key] = "test"
        req_params["dictionary"] = json.dumps(dic)

        res = test_session.put(url, req_params)
        time += (res.elapsed.seconds * 1000000) + res.elapsed.microseconds

        if res.status_code != 200:
            break

    stats['batch_time'] = round((time / (len(keys) / batch_size)) / 1000)
    stats['batch_total'] = round(time / 1000)
    stats['batch_size'] = batch_size
    stats['time'] += time


def test_service(size, distance, results):
    """ We evaluate the performance of the service for a specific number of
        keys in the service, while querying with a user-defined maximum distance
        and number of results.

        Args:
            size (int): the size of the testfile dataset. File should be
                        generated before by using the script provided in test
                        folder.
            distance (int): the distance until where we span our searches
            results (int): maximum number of results which the query will return

        Returns:
            string: it represents the json encoding of some stats
    """
    # We generate a random id for the test store in order to avoid conflicts
    store_name = uuid.uuid4()

    # We read data from the test file
    keys, queries = read_data(
        "data/testset_{0}.dat".format(size))

    # We will hold the stats in this dictionary
    stats = {
        'time': 0,
        'keys': len(keys),
        'queries': len(queries),
        'distance': distance,
        'results': results,
    }
    # We create a session and create a new test database with that session
    test_session = requests.Session()
    create_store(test_session, store_name)

    # We upload the keys to the service
    upload_keys(test_session, store_name, keys, stats)

    # Execute queries and benchmark service

    res = test_session.get('http://localhost:8080/fuzzy/batch',
                           params={
                               'store': store_name,
                               'distance': distance,
                               'results': results,
                               'keys': json.dumps(
                                   [query[1] for query in queries])
                           })
    response_time = (res.elapsed.microseconds / 1000) + \
        (res.elapsed.seconds * 1000)
    stats['time'] += response_time

    accuracy = 0
    for correct, response in zip([query[0] for query in queries],
                                 json.loads(res.text)):
        try:
            index = response.index(correct)
            accuracy += (len(response) - index) / len(response)
        except ValueError:
            pass

    stats['accuracy'] = round((accuracy * 100) / len(queries), 2)
    stats['time'] = round(stats['time'])
    stats['throughput'] = round((1000 / (response_time)) * len(queries))

    delete_store(test_session, store_name)

    return json.dumps(stats)
