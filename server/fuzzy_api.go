package server

import (
	"../fuzzy"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store", "key", "distance"}, w, r)
	if !valid {
		return
	}

	store, present := getStore(parameters["store"])
	if !present {
		parameterError(w, "store")
		return
	}

	/* We always require a distance parameter in order to make every request more explicit
	   about whether we would like to perform and exact match or an approximate one */

	distance, err := strconv.Atoi(parameters["distance"])
	if err != nil {
		parameterError(w, "distance (numeric)")
		return
	}

	/* We treat exact matching here */
	if distance == 0 {
		value, present := store.Get(parameters["key"])
		if !present {
			parameterError(w, "key (Existent)")
			return
		}
		incrementStats(parameters["store"], "/fuzzy GET")
		fmt.Fprintf(w, value)
		return
	}

	/* Approximate matching here */
	results, err := strconv.Atoi(r.FormValue("results"))
	if err != nil {
		parameterError(w, "results")
		return
	}
	fuzzyResults := store.Query(parameters["key"], distance, results)
	jsonResponse, _ := json.Marshal(fuzzyResults)
	incrementStats(parameters["store"], "/fuzzy GET")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))

}

func newStoreHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	_, present := getStore(parameters["store"])
	if present {
		http.Error(w, "This store already exists", http.StatusBadRequest)
		return
	}

	fuzzyStore.stores[parameters["store"]] = fuzzy.NewService()

	fuzzyStore.StatsLock.Lock()
	fuzzyStore.stats[parameters["store"]] = storeStatistics{make(map[string]int)}
	fuzzyStore.StatsLock.Unlock()
	incrementStats(parameters["store"], "/fuzzy POST")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Store has successfully been created")

}

func addKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store", "key", "value"}, w, r)
	if !valid {
		return
	}

	store, present := getStore(parameters["store"])
	if !present {
		parameterError(w, "store (existent)")
		return
	}

	fuzzyStore.StoresLock.Lock()
	store.Set(parameters["key"], parameters["value"])
	fuzzyStore.StoresLock.Unlock()

	fmt.Fprintf(w, "Successfully set the key")
	incrementStats(parameters["store"], "/fuzzy PUT")
}

func deleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	store, present := getStore(parameters["store"])
	if !present {
		parameterError(w, "store (existent)")
		return
	}
	/* If there is no key parameter delete the entire collection */
	key := r.FormValue("key")
	if len(key) == 0 {

		fuzzyStore.StoresLock.Lock()
		delete(fuzzyStore.stores, parameters["store"])
		fuzzyStore.StoresLock.Unlock()

		fuzzyStore.StatsLock.Lock()
		delete(fuzzyStore.stats, parameters["store"])
		fuzzyStore.StatsLock.Unlock()

		return
	}

	deleted := store.Delete(parameters["key"])
	if deleted {
		fmt.Fprintf(w, "Successfully deleted the key")
		incrementStats(parameters["store"], "/fuzzy DELETE")
	} else {
		http.Error(w, "The key you deleted does not exist", http.StatusBadRequest)
	}
}

// FuzzyHandler handles all requests which operate on a single key or value.
// It follows the API described in the project's wiki.
func FuzzyHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getKeyHandler(w, r)
		return
	case "PUT":
		addKeyValueHandler(w, r)
		return
	case "DELETE":
		deleteKeyHandler(w, r)
		return
	case "POST":
		newStoreHandler(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
