package server

import (
	"../fuzzy"
	"fmt"
	"net/http"
)

func parameterError(w http.ResponseWriter, parameter string) {
	http.Error(w, fmt.Sprintf("Please provide a valid %s parameter", parameter), http.StatusBadRequest)
}

func requireParameters(parameters []string, w http.ResponseWriter, r *http.Request) (map[string]string, bool) {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter] = r.FormValue(parameter)
		if len(result[parameter]) == 0 {
			parameterError(w, parameter)
			return result, false
		}
	}
	return result, true
}

func incrementStats(store, operation string) {
	fuzzyStore.StatsLock.Lock()
	fuzzyStore.stats[store].Queries[operation]++
	fuzzyStore.StatsLock.Unlock()
}

func getStore(name string) (*fuzzy.Service, bool) {
	fuzzyStore.StoresLock.RLock()
	store, present := fuzzyStore.stores[name]
	fuzzyStore.StoresLock.RUnlock()
	return store, present
}
