package service

import (
	"container/heap"
	"github.com/ryszard/goskiplist/skiplist"
	"github.com/vlad-doru/fuzzyguy/levenshtein"
	"sync"
)

type Service interface {
	Add(key, value string)
	Get(key string) (string, bool)
	Query(key string, distance, max_results int) []string
	Len() int
}

type Storage struct {
	key      string
	value    string
	extended uint64
}

type FuzzyService struct {
	dictionary map[int]map[uint32][]Storage
	histograms map[int]skiplist.SkipList
}

func NewFuzzyService() FuzzyService {
	dict := make(map[int]map[uint32][]Storage)
	histo := make(map[int]skiplist.SkipList)
	return FuzzyService{dict, histo}
}

func (service FuzzyService) Add(key, value string) {
	histogram := levenshtein.ComputeHistogram(key)
	storage := Storage{key, value, levenshtein.ComputeExtendedHistogram(key)}
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogram_present := bucket[histogram]
		if histogram_present {
			for _, pair := range list {
				if pair.key == key {
					pair.value = value
					return
				}
			}
			bucket[histogram] = append(bucket[histogram], storage)
			return
		}
		bucket[histogram] = []Storage{storage}
		return
	}
	bucket = map[uint32][]Storage{histogram: []Storage{storage}}
	service.dictionary[len(key)] = bucket
}

func (service FuzzyService) Get(key string) (string, bool) {
	histogram := levenshtein.ComputeHistogram(key)
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogram_present := bucket[histogram]
		if histogram_present {
			for _, pair := range list {
				if pair.key == key {
					return pair.value, true
				}
			}
		}
	}
	return "", false
}

func (service FuzzyService) Len() int {
	result := 0
	for _, dict := range service.dictionary {
		result += len(dict)
	}
	return result
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
	query_extended := levenshtein.ComputeExtendedHistogram(query)
	query_len := len(query)
	mutex := sync.RWMutex{}
	sync_channel := make(chan int)

	for i := query_len - threshold; i < query_len+threshold+1; i++ {
		diff := Abs(i - query_len)
		go func(dict map[uint32][]Storage) {
			for histogram, list := range dict {
				if levenshtein.LowerBound(query_histogram, histogram, diff) > threshold {
					continue
				}
				for _, pair := range list {
					if levenshtein.ExtendedLowerBound(query_extended, pair.extended, diff) > threshold {
						continue
					}
					distance, within := levenshtein.DistanceThreshold(query, pair.key, threshold)
					if within {
						mutex.Lock()
						heap.Push(h, KeyScore{distance, pair.key})
						if h.Len() > max_results {
							heap.Pop(h)
						}
						mutex.Unlock()
					}
				}
			}
			sync_channel <- 1
		}(service.dictionary[i])
	}
	for i := query_len - threshold; i < query_len+threshold+1; i++ {
		<-sync_channel
	}

	results := make([]string, h.Len())
	for i := 0; i < len(results); i++ {
		results[i] = h.Pop().(KeyScore).key
	}
	return results
}
