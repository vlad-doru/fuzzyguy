package service

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestKeyScoreHeap(t *testing.T) {
	h := new(KeyScoreHeap)
	heap.Init(h)
	heap.Push(h, KeyScore{key: "test", score: 1})
	heap.Push(h, KeyScore{key: "test", score: 0})
	heap.Push(h, KeyScore{key: "test", score: 3})
	heap.Push(h, KeyScore{key: "test", score: 4})

	order := []int{4, 3, 1, 0}

	for _, val := range order {
		popped := heap.Pop(h).(KeyScore)
		if val != popped.score {
			t.Log(popped)
			t.Error("Heap does not work properly")
		}
	}
}

func BenchmarkHeapOperations(b *testing.B) {
	h := new(KeyScoreHeap)
	heap.Init(h)
	for n := 0; n < b.N; n++ {
		heap.Push(h, KeyScore{1, "test"})
		if h.Len() > 5 {
			heap.Pop(h)
		}
	}
}

func TestSimpleAddGet(t *testing.T) {
	service := NewFuzzyService()
	service.Add("key", "value")
	value, _ := service.Get("key")
	if value != "value" {
		t.Log(value)
		t.Error("Simple add & get test fails")
	}
	_, present := service.Get("nonexistent")
	if present {
		t.Error("Get of nonexistent fails")
	}
}

func BenchmarkServiceAdd(b *testing.B) {
	service := NewFuzzyService()
	for n := 0; n < b.N; n++ {
		service.Add("key", "value")
	}
}

func BenchmarkServiceGet(b *testing.B) {
	service := NewFuzzyService()
	service.Add("key", "value")
	for n := 0; n < b.N; n++ {
		service.Get("key")
	}
}

func TestFuzzyService(t *testing.T) {
	service := NewFuzzyService()
	service.Add("ana", "super")
	service.Add("anan", "value")
	service.Add("super", "ceva")
	service.Add("supret", "altceva")
	service.Add("supretar", "altceva")

	result := service.Query("supre", 2, 1)
	if result[0] != "supret" {
		t.Log("supre Query wrong")
		t.Error("Failed the query action")
	}

	result = service.Query("supre", 3, 2)
	if result[0] != "supret" {
		t.Log("supre Query wrong")
		t.Error("Failed the query action")
	}
	if result[1] != "super" {
		t.Log("supre Query wrong")
		t.Error("Failed the query action")
	}
}

func LoadTestSet(name string) ([]string, []string, *FuzzyService) {
	service := NewFuzzyService()

	var key string
	var keys_nr, queries_nr int
	var queries, correct []string

	fmt.Printf("[FILE NAME %s ]\n", name)
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(file)

	l, _, _ := reader.ReadLine()
	keys_nr, _ = strconv.Atoi(string(l))

	for i := 0; i < keys_nr; i++ {
		l, _, _ = reader.ReadLine()
		key = string(l)
		service.Add(key, "test")
	}

	l, _, _ = reader.ReadLine()
	queries_nr, _ = strconv.Atoi(string(l))

	queries = make([]string, queries_nr)
	correct = make([]string, queries_nr)

	for i := 0; i < queries_nr; i++ {
		l, _, _ = reader.ReadLine()
		split := strings.Split(string(l), "\t")
		correct[i], queries[i] = split[0], split[1]
	}
	return queries, correct, service
}

func BenchmarkParallelServiceQuery(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())

	queries, correct, service := LoadTestSet("data/testset_300000.dat")

	var accuracy float32 = 0
	c := make(chan float32)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		accuracy = 0
		for qindex, _ := range queries {
			go func(j int) {
				result := service.Query(queries[j], 3, 5)
				var precision float32 = 0
				for rindex, res := range result {
					if res == correct[j] {
						precision = float32(rindex+1) / float32(len(result))
						break
					}
				}
				c <- precision
			}(qindex)

		}
		for j := 0; j < len(queries); j++ {
			accuracy += <-c
		}
	}
	fmt.Printf("Accuracy of testset \t %f\n------------\n", accuracy/float32(len(queries)))

}

func BenchmarkSequentialServiceQuery(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())

	queries, correct, service := LoadTestSet("data/testset_300000.dat")
	var accuracy float32 = 0

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for qindex, _ := range queries {
			result := service.Query(queries[qindex], 3, 5)
			b.StopTimer()
			for rindex, res := range result {
				if res == correct[qindex] {
					accuracy += float32(rindex+1) / float32(len(result))
					break
				}
			}
			b.StartTimer()
		}
		b.StopTimer()
		fmt.Printf("Accuracy of testset \t %f\n------------\n", accuracy/float32(len(queries)))
		b.StartTimer()
	}
}
