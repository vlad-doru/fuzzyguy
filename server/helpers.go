package server

import (
	"fmt"
	"../fuzzy"
	"net/http"
)

func ParameterError(w http.ResponseWriter, parameter string) {
	http.Error(w, fmt.Sprintf("Please provide a valid %s parameter", parameter), http.StatusBadRequest)
}

func RequireParameters(parameters []string, w http.ResponseWriter, r *http.Request) (map[string]string, bool) {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter] = r.FormValue(parameter)
		if len(result[parameter]) == 0 {
			ParameterError(w, parameter)
			return result, false
		}
	}
	return result, true
}

func IncrementStats(store, operation string) {
	Fuzzy.stats_lock.Lock()
	Fuzzy.stats[store].Queries[operation] += 1
	Fuzzy.stats_lock.Unlock()
}

func GetStore(name string) (*fuzzy.FuzzyService, bool) {
	Fuzzy.stores_lock.RLock()
	store, present := Fuzzy.stores[name]
	Fuzzy.stores_lock.RUnlock()
	return store, present
}
