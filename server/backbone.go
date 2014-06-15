package server

import (
	"../fuzzy"
	"sync"
)

type StoreStatistics struct {
	Queries map[string]int
}

type Statistics struct {
	Stores map[string]StoreStatistics
	mutex  *sync.RWMutex
}

type Server struct {
	stores      map[string]*fuzzy.FuzzyService
	stats       map[string]StoreStatistics
	stats_lock  sync.RWMutex
	stores_lock sync.RWMutex
}

var Fuzzy = Server{
	stores:      make(map[string]*fuzzy.FuzzyService),
	stats:       make(map[string]StoreStatistics),
	stats_lock:  sync.RWMutex{},
	stores_lock: sync.RWMutex{}}
