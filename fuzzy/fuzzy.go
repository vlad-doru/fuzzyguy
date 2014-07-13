// Package fuzzy defines the necessary tools to manipulate data
//
// Exported types:
// Service -> data structure to encapsulate all information
//
// Exported functions:
// NewService -> constructor function.
// Get, Set, Len, Query -> methods for the Service type.
package fuzzy

import (
	"../levenshtein"
	"container/heap"
	"sort"
	"sync"
)

type service interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Query(key string, distance, maxResults int) []string
	Len() int
}

type storage struct {
	key      string
	value    string
	extended uint64
}

// Service is a type which holds the underlying data structures used
// to implement the fuzzy service.
//
// Attributes:
// dictionary (map[int]map[uint32][]storage): Maps each storage type
// (which encapsulates necessary data) to its corresponding 32bit histogram
// and the length of the key indexed
//
// rwmutex (*sync.RWMutex): Read-write mutex used to synchronize operations on
// the dictionary without having data races
type Service struct {
	dictionary map[int]map[uint32][]storage
	rwmutex    *sync.RWMutex
}

// NewService is a constructor function for a Service object.
//
// Arugments: None
//
// Returns: (*Service) a pointer to an object which can then be used
// to manipulate data through the Set/Get/Query operations.
func NewService() *Service {
	dict := make(map[int]map[uint32][]storage)
	mutex := &sync.RWMutex{}
	return &Service{dict, mutex}
}

// Set a value to be indexed by a specific key in the system
//
// Arguments:
// key (string): the key to index the specific value
// value (string): the value to be indexed by the specified key
func (service Service) Set(key, value string) {
	histogram := levenshtein.ComputeHistogram(key)
	storeValue := storage{key, value, levenshtein.ComputeExtendedHistogram(key)}
	service.rwmutex.Lock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogramPresent := bucket[histogram]
		if histogramPresent {
			for i, pair := range list {
				if pair.key == key {
					list[i].value = value
					service.rwmutex.Unlock()
					return
				}
			}
			bucket[histogram] = append(bucket[histogram], storeValue)
		} else {
			bucket[histogram] = []storage{storeValue}
		}
	} else {
		bucket = map[uint32][]storage{histogram: {storeValue}}
		service.dictionary[len(key)] = bucket
	}
	service.rwmutex.Unlock()
}

// Delete a key from the system.
//
// Arguments:
// key (string): the key to be deleted from the system along with the values
// 	which it indexes.
//
// Returns: (bool) wether the deletion was susccesful or not
func (service Service) Delete(key string) bool {
	histogram := levenshtein.ComputeHistogram(key)
	service.rwmutex.Lock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogramPresent := bucket[histogram]
		if histogramPresent {
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

// Get the value associated with a specific key
//
// Arguments:
// key (string): the key whose value which we want to return
//
// Returns: (string, bool) tuple which represents (value, present). If present
// is false then the value we return is the empty string ""
func (service Service) Get(key string) (string, bool) {
	histogram := levenshtein.ComputeHistogram(key)
	service.rwmutex.RLock()
	bucket, present := service.dictionary[len(key)]
	if present {
		list, histogramPresent := bucket[histogram]
		if histogramPresent {
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

// Len returns the number of keys indexed by the system
//
// Returns: (int) the total number of keys indexed by the system
func (service Service) Len() int {
	result := 0
	service.rwmutex.RLock()
	for _, dict := range service.dictionary {
		result += len(dict)
	}
	service.rwmutex.RUnlock()
	return result
}

// TODO: Move the keyScore, keyScoreHeap implementation to a different file
type keyScore struct {
	prefix int
	score  int
	key    string
}

type keyScoreHeap []keyScore

func (h keyScoreHeap) Len() int {
	return len(h)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func prefix(source, target string) int {
	prefix := 0
	for i := 0; i < min(len(source), len(target)); i++ {
		if source[i] == target[i] {
			prefix++
		} else {
			break
		}
	}
	return prefix
}

// We implement the heap methods here: Less, Swap, Push, Pop

func (h keyScoreHeap) Less(i, j int) bool {
	if h[i].prefix != h[j].prefix {
		return h[i].prefix < h[j].prefix
	}
	if h[i].score != h[j].score {
		return h[i].score > h[j].score
	}
	return h[i].key > h[j].key // this is the max-heap condition
}

func (h keyScoreHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *keyScoreHeap) Push(x interface{}) {
	*h = append(*h, x.(keyScore))
}

func (h *keyScoreHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Query the service for keys which can have a Levenshtein distance smaller
// than a threshold for a spefici key. Take only the first x result
//
// Arugments:
// query (string): the base key
// threshold (int): how far can a candidate be in the Levenshtein metric space
// maxResults (int): the maximum number of results which will be returned
func (service Service) Query(query string, threshold, maxResults int) []string {
	h := new(keyScoreHeap)
	heap.Init(h)
	queryHistogram := levenshtein.ComputeHistogram(query)
	queryExtended := levenshtein.ComputeExtendedHistogram(query)
	queryLen := len(query)
	heapMutex := &sync.Mutex{}
	syncChannel := make(chan int)
	start := queryLen - threshold
	stop := queryLen + threshold + 1

	service.rwmutex.RLock()
	for i := start; i < stop; i++ {
		go func(index int, mutex *sync.Mutex) {
			diff := abs(index - queryLen)
			for histogram, list := range service.dictionary[index] {
				if levenshtein.LowerBound(queryHistogram, histogram, diff) > threshold {
					continue
				}
				for _, pair := range list {
					if levenshtein.ExtendedLowerBound(queryExtended, pair.extended, diff) > threshold {
						continue
					}
					distance, within := levenshtein.DistanceThreshold(query, pair.key, threshold)
					if within {
						mutex.Lock()
						heap.Push(h, keyScore{prefix(pair.key, query), distance, pair.key})
						if h.Len() > maxResults {
							heap.Pop(h)
						}
						mutex.Unlock()
					}
				}
			}
			syncChannel <- 1
		}(i, heapMutex)
	}
	for i := start; i < stop; i++ {
		<-syncChannel
	}
	service.rwmutex.RUnlock()

	sort.Sort(h)
	results := make([]string, h.Len())
	for i := 0; i < len(results); i++ {
		results[i] = h.Pop().(keyScore).key
	}
	return results
}
