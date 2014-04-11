package service

import (
	"container/heap"
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
