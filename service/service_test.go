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