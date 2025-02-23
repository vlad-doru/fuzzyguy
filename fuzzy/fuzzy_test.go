package fuzzy

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

func TestkeyScoreHeap(t *testing.T) {
	h := new(keyScoreHeap)
	heap.Init(h)
	heap.Push(h, keyScore{key: "test", score: 1})
	heap.Push(h, keyScore{key: "test", score: 0})
	heap.Push(h, keyScore{key: "test", score: 3})
	heap.Push(h, keyScore{key: "test", score: 4})

	order := []int{4, 3, 1, 0}

	for _, val := range order {
		popped := heap.Pop(h).(keyScore)
		if val != popped.score {
			t.Log(popped)
			t.Error("Heap does not work properly")
		}
	}
}

func BenchmarkHeapOperations(b *testing.B) {
	h := new(keyScoreHeap)
	heap.Init(h)
	for n := 0; n < b.N; n++ {
		heap.Push(h, keyScore{score: 1, key: "test"})
		if h.Len() > 5 {
			heap.Pop(h)
		}
	}
}

func TestSimpleSetGetDelete(t *testing.T) {
	service := NewService()
	service.Set("key", "value")
	value, _ := service.Get("key")
	if value != "value" {
		t.Log(value)
		t.Error("Simple Set & get test fails")
	}
	_, present := service.Get("nonexistent")
	if present {
		t.Error("Get of nonexistent fails")
	}
	service.Delete("key")
	value, present = service.Get("key")
	if present {
		t.Error("Get of nonexistent after delete fails")
		t.Error(value)
	}
	service.Set("key", "test")
	service.Set("kye", "test")
	service.Delete("key")
	service.Delete("nonexistent")
	service.Set("another", "test")
	value, present = service.Get("key")
	if present {
		t.Error("Get of nonexistent after delete fails")
		t.Error(value)
	}

	service.Set("key", "test")
	value, _ = service.Get("key")
	if value != "test" {
		t.Error("Service did not record our change")
		t.Error(value)
	}

	service.Set("key", "another")
	value, _ = service.Get("key")
	if value != "another" {
		t.Error("Service did not record our change")
		t.Error(value)
	}
}

func BenchmarkServiceSet(b *testing.B) {
	service := NewService()
	for n := 0; n < b.N; n++ {
		service.Set("key", "value")
	}
}

func BenchmarkServiceGet(b *testing.B) {
	service := NewService()
	service.Set("key", "value")
	for n := 0; n < b.N; n++ {
		service.Get("key")
	}
}

func TestService(t *testing.T) {
	service := NewService()
	service.Set("ana", "super")
	service.Set("anan", "value")
	service.Set("super", "ceva")
	service.Set("supret", "altceva")
	service.Set("supretar", "altceva")

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
	if result[1] != "supretar" {
		t.Log("supre Query wrong")
		t.Error("Failed the query action")
	}
}

const testFile = "../test/data/testset_300000.dat"

func TestConcurrencyService(t *testing.T) {
	queries, _, service := LoadTestSet(testFile)
	service = NewService()

	barrier := make(chan bool)
	steps := 1000
	go func() {
		for i := 0; i < steps; i++ {
			<-barrier
		}
	}()
	for i := 0; i < steps; i++ {
		go func(index int) {
			service.Set(queries[index], "test")
			barrier <- true
		}(i % len(queries))
		service.Set("another", "test")
	}
}

func LoadTestSet(name string) ([]string, []string, *Service) {
	service := NewService()

	var key string
	var keysNr, queriesNr int
	var queries, correct []string

	fmt.Printf("\n[FILE NAME %s ]\n", name)
	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(file)

	l, _, _ := reader.ReadLine()
	keysNr, _ = strconv.Atoi(string(l))

	for i := 0; i < keysNr; i++ {
		l, _, _ = reader.ReadLine()
		key = string(l)
		service.Set(key, "test")
	}

	l, _, _ = reader.ReadLine()
	queriesNr, _ = strconv.Atoi(string(l))

	queries = make([]string, queriesNr)
	correct = make([]string, queriesNr)

	for i := 0; i < queriesNr; i++ {
		l, _, _ = reader.ReadLine()
		split := strings.Split(string(l), "\t")
		correct[i], queries[i] = split[0], split[1]
	}
	return queries, correct, service
}

func BenchmarkParallelServiceQuery(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())

	queries, _, service := LoadTestSet(testFile)

	c := make(chan bool)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		for _, query := range queries {
			go func(query string) {
				service.Query(query, 3, 5)
				c <- true
			}(query)
		}
		for j := 0; j < len(queries); j++ {
			<-c
		}
	}
}

func BenchmarkAccuracyServiceQuery(b *testing.B) {
	b.StopTimer()
	runtime.GOMAXPROCS(runtime.NumCPU())

	queries, correct, service := LoadTestSet(testFile)

	var accuracy float32
	c := make(chan float32)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		accuracy = 0
		for qindex := range queries {
			go func(j int) {
				result := service.Query(queries[j], 3, 5)
				var precision float32
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

	queries, correct, service := LoadTestSet(testFile)
	var accuracy float32

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for qindex := range queries {
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
