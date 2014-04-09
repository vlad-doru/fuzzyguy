package service

import (
	//"container/heap"
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

func (service FuzzyService) Add(key, value string) {
	service.dictionary[len(key)][key] = Storage{levenshtein.ComputeHistogram(key), value}
}

func (service FuzzyService) Get(key string) (string, bool) {
	storage, present := service.dictionary[len(key)][key]
	if present {
		return storage.value, true
	} else {
		return "", false
	}
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
	return h[j].score < h[i].score // this is the max-heap condition
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

func (service FuzzyService) Query(key string, distance, max_results int) []string {
	heap := new(KeyScoreHeap)
	// query_histogram := levenshtein.ComputeHistogram(key)

	results := make([]string, max_results)
	for i, result := range *heap {
		results[i] = result.key
	}
	return results
}
