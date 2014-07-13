package server

import (
	"../fuzzy"
	"sync"
)

type storeStatistics struct {
	Queries map[string]int
}

type statistics struct {
	Stores map[string]storeStatistics
	mutex  *sync.RWMutex
}

type server struct {
	stores     map[string]*fuzzy.Service
	stats      map[string]storeStatistics
	StatsLock  sync.RWMutex
	StoresLock sync.RWMutex
}

var fuzzyStore = server{
	stores:     make(map[string]*fuzzy.Service),
	stats:      make(map[string]storeStatistics),
	StatsLock:  sync.RWMutex{},
	StoresLock: sync.RWMutex{}}
