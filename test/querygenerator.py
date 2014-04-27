from __future__ import print_function

import numpy
import string

from progressbar import ProgressBar
from collections import defaultdict
from random import sample, choice, gauss, randint, sample
from source.stats.discrete_generator import discrete_gen


def load_vocabulary(filename):
    """We return an immutable set with the vocabulary in it"""
    vocabulary = defaultdict(list)

    with open(filename) as f:
        lines = f.readlines()

        for line in lines:
            word = line.strip()
            vocabulary[len(word)].append(word)
    return vocabulary


def alter_queries(queries, f):

    randchar = lambda: choice(string.ascii_letters.lower())

    def insert(x):
        index = randint(0, len(x))
        return x[:index] + randchar() + x[index:]

    def delete(x):
        index = randint(0, len(x) - 1)
        return x[:index] + x[index + 1:]

    def subsititute(x):
        index = randint(0, len(x) - 1)
        return x[:index] + randchar() + x[index + 1:]

    def transposition(x):
        index = randint(0, len(x) - 2)
        return x[:index] + x[index + 1] + x[index] + x[index + 2:]

    distance_probs = {
        1: 0.63,
        2: 0.21,
        3: 0.16,
    }

    mistakes_probs = {
        insert: 0.44,
        delete: 0.28,
        subsititute: 0.26,
        transposition: 0.02
    }

    distance_gen = discrete_gen(distance_probs)
    mistake_gen = discrete_gen(mistakes_probs)

    mistake = next(mistake_gen)

    nr_queries = 1000
    size = len(queries)

    print(nr_queries, file=f)
    print("Generating alterated queries")

    pbar = ProgressBar(maxval=nr_queries).start()
    for i in range(nr_queries):
        # index = round(abs(gauss(0, 1)) * size) % size
        index = randint(0, size)
        query = queries[index]
        distance, misspelled = next(distance_gen), query
        for step in range(distance):
            mistake = next(mistake_gen)
            misspelled = mistake(misspelled)
        print(query + '\t' + misspelled, file=f)
        pbar.update(i)
    pbar.finish()


def generate_queries(vocabulary, size):
    lengths = numpy.abs(numpy.random.normal(10, 5, size))
    words_set = []

    filename = "data/testset_{0}.dat".format(size)
    print("Generating queries for size {0}".format(size))
    f = open(filename, 'w')
    print(size, file=f)

    for _, words in vocabulary.items():
        for word in words:
            words_set.append(word)

    random_word = lambda: choice(words_set)
    gen = set()

    pbar = ProgressBar(maxval=size).start()
    for index, length in enumerate(lengths):
        while True:
            query = random_word()
            while len(query) < length - 2:
                if (length - len(query) - 1) in vocabulary:
                    query += " " + choice(vocabulary[length - len(query) - 1])
                else:
                    word = random_word()
                    while len(query) + len(word) > length + 2:
                        word = random_word()
                    query += (" " + word)
            if query in gen:
                continue
            gen.add(query)
            print(query, file=f)
            break
        pbar.update(index)
    pbar.finish()

    gen = list(gen)
    alter_queries(gen, f)

    f.close()


def main():
    vocabulary = load_vocabulary("data/english.txt")
    for size in [5000, 50000, 100000, 200000, 300000]:
        generate_queries(vocabulary, size)

if __name__ == '__main__':
    main()
