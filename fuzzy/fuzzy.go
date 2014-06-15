package fuzzy

import (
	"container/heap"
	"../levenshtein"
	"sort"
	"sync"
)

type Service interface {
	Set(key, value string)
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
	rwmutex    *sync.RWMutex
}

func NewFuzzyService() *FuzzyService {
	dict := make(map[int]map[uint32][]Storage)
	mutex := &sync.RWMutex{}
	return &FuzzyService{dict, mutex}
}

func (service FuzzyService) Set(key, value string) {
	histogram := levenshtein.ComputeHistogram(key)
	storage := Storage{key, value, levenshtein.ComputeExtendedHistogram(key)}
	service.rwmutex.Lock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogram_present := bucket[histogram]
		if histogram_present {
			for i, pair := range list {
				if pair.key == key {
					list[i].value = value
					service.rwmutex.Unlock()
					return
				}
			}
			bucket[histogram] = append(bucket[histogram], storage)
		} else {
			bucket[histogram] = []Storage{storage}
		}
	} else {
		bucket = map[uint32][]Storage{histogram: []Storage{storage}}
		service.dictionary[len(key)] = bucket
	}
	service.rwmutex.Unlock()
}

func (service FuzzyService) Delete(key string) bool {
	histogram := levenshtein.ComputeHistogram(key)
	service.rwmutex.Lock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogram_present := bucket[histogram]
		if histogram_present {
			for index, pair := range list {
				if pair.key == key {
					list[index], list = list[len(list)-1], list[:len(list)-1]
					if len(list) == 0 {
						delete(bucket, histogram)
						if len(bucket) == 0 {
							delete(service.dictionary, len(key))
						}
						service.rwmutex.Unlock()
						return true
					}
					bucket[histogram] = list
					service.rwmutex.Unlock()
					return true
				}
			}
		}
	}
	service.rwmutex.Unlock()
	return false
}

func (service FuzzyService) Get(key string) (string, bool) {
	histogram := levenshtein.ComputeHistogram(key)
	service.rwmutex.RLock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogram_present := bucket[histogram]
		if histogram_present {
			for _, pair := range list {
				if pair.key == key {
					service.rwmutex.RUnlock()
					return pair.value, true
				}
			}
		}
	}
	service.rwmutex.RUnlock()
	return "", false
}

func (service FuzzyService) Len() int {
	result := 0
	service.rwmutex.RLock()
	for _, dict := range service.dictionary {
		result += len(dict)
	}
	service.rwmutex.RUnlock()
	return result
}

type KeyScore struct {
	prefix int
	score  int
	key    string
}

type KeyScoreHeap []KeyScore

// We are going to implement a KeyScore max-heap based on the score
func (h KeyScoreHeap) Len() int {
	return len(h)
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Prefix(source, target string) int {
	prefix := 0
	for i := 0; i < Min(len(source), len(target)); i++ {
		if source[i] == target[i] {
			prefix++
		} else {
			break
		}
	}
	return prefix
}

func (h KeyScoreHeap) Less(i, j int) bool {
	if h[i].prefix != h[j].prefix {
		return h[i].prefix < h[j].prefix
	}
	if h[i].score != h[j].score {
		return h[i].score > h[j].score
	}
	return h[i].key > h[j].key // this is the max-heap condition
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
	heap_mutex := &sync.Mutex{}
	sync_channel := make(chan int)
	start := query_len - threshold
	stop := query_len + threshold + 1

	service.rwmutex.RLock()
	for i := start; i < stop; i++ {
		go func(index int, mutex *sync.Mutex) {
			diff := Abs(index - query_len)
			for histogram, list := range service.dictionary[index] {
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
						heap.Push(h, KeyScore{Prefix(pair.key, query), distance, pair.key})
						if h.Len() > max_results {
							heap.Pop(h)
						}
						mutex.Unlock()
					}
				}
			}
			sync_channel <- 1
		}(i, heap_mutex)
	}
	for i := start; i < stop; i++ {
		<-sync_channel
	}
	service.rwmutex.RUnlock()

	sort.Sort(h)
	results := make([]string, h.Len())
	for i := 0; i < len(results); i++ {
		results[i] = h.Pop().(KeyScore).key
	}
	return results
}
