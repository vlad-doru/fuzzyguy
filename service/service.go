package service

import (
	"container/heap"
	"github.com/vlad-doru/fuzzyguy/levenshtein"
)

type Service interface {
	Add(key, value string)
	Get(key string) (string, bool)
	Query(key string, distance, max_results int) []string
	Len() int
}

type Storage struct {
	histogram uint32
	value     string
}

type FuzzyService struct {
	dictionary map[int]map[string]Storage
}

func NewFuzzyService() FuzzyService {
	dict := make(map[int]map[string]Storage)
	return FuzzyService{dict}
}

func (service FuzzyService) Add(key, value string) {
	storage := Storage{levenshtein.ComputeHistogram(key), value}
	bucket, present := service.dictionary[len(key)]
	if present {
		bucket[key] = storage
	} else {
		bucket = map[string]Storage{key: storage}
		service.dictionary[len(key)] = bucket
	}
}

func (service FuzzyService) Get(key string) (string, bool) {
	bucket, present := service.dictionary[len(key)]
	if present {
		_, present = bucket[key]
		if present {
			return bucket[key].value, true
		}
	}
	return "", false
}

func (service FuzzyService) Len() int {
	return len(service.dictionary)
}

type KeyScore struct {
	score int
	key   string
}

type KeyScoreHeap []KeyScore

// We are going to implement a KeyScore max-heap based on the score
func (h KeyScoreHeap) Len() int {
	return len(h)
}

func (h KeyScoreHeap) Less(i, j int) bool {
	return h[i].score > h[j].score // this is the max-heap condition
}

func (h KeyScoreHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *KeyScoreHeap) Push(x interface{}) {
	*h = append(*h, x.(KeyScore))
}

func (h *KeyScoreHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (service FuzzyService) Query(query string, threshold, max_results int) []string {
	h := new(KeyScoreHeap)
	heap.Init(h)
	query_histogram := levenshtein.ComputeHistogram(query)
	query_len := len(query)

	for i := query_len - threshold; i < query_len+threshold+1; i++ {
		diff := Abs(i - query_len)
		for key, storage := range service.dictionary[i] {
			if levenshtein.LowerBound(query_histogram, storage.histogram, diff) <= threshold {
				distance, within := levenshtein.DistanceThreshold(query, key, threshold)
				if within {
					if distance <= threshold {
						heap.Push(h, KeyScore{distance, key})
						if h.Len() > max_results {
							heap.Pop(h)
						}
					}
				}
			}
		}
	}

	results := make([]string, h.Len())
	for i := 0; i < len(results); i++ {
		results[i] = h.Pop().(KeyScore).key
	}
	return results
}
