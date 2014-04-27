package server

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"net/http"
	"strconv"
)

func GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "key", "distance"}, w, r)
	if !valid {
		return
	}

	store, present := GetStore(parameters["store"])
	if !present {
		ParameterError(w, "store")
		return
	}

	/* We always require a distance parameter in order to make every request more explicit
	   about whether we would like to perform and exact match or an approximate one */

	distance, err := strconv.Atoi(parameters["distance"])
	if err != nil {
		ParameterError(w, "distance (numeric)")
		return
	}

	/* We treat exact matching here */
	if distance == 0 {
		value, present := store.Get(parameters["key"])
		if !present {
			ParameterError(w, "key (Existent)")
			return
		}
		IncrementStats(parameters["store"], "/fuzzy GET")
		fmt.Fprintf(w, value)
		return
	}

	/* Approximate matching here */
	results, err := strconv.Atoi(r.FormValue("results"))
	if err != nil {
		ParameterError(w, "results")
		return
	}
	fuzzy_results := store.Query(parameters["key"], distance, results)
	json_response, _ := json.Marshal(fuzzy_results)
	IncrementStats(parameters["store"], "/fuzzy GET")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json_response))

}

func NewStoreHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	_, present := GetStore(parameters["store"])
	if present {
		http.Error(w, "This store already exists", http.StatusBadRequest)
		return
	}

	Fuzzy.stores[parameters["store"]] = fuzzy.NewFuzzyService()

	Fuzzy.stats_lock.Lock()
	Fuzzy.stats[parameters["store"]] = StoreStatistics{make(map[string]int)}
	Fuzzy.stats_lock.Unlock()
	IncrementStats(parameters["store"], "/fuzzy POST")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Store has successfully been created")

}

func AddKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "key", "value"}, w, r)
	if !valid {
		return
	}

	store, present := GetStore(parameters["store"])
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	Fuzzy.stores_lock.Lock()
	store.Set(parameters["key"], parameters["value"])
	Fuzzy.stores_lock.Unlock()

	fmt.Fprintf(w, "Successfully set the key")
	IncrementStats(parameters["store"], "/fuzzy PUT")
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	store, present := GetStore(parameters["store"])
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	/* If there is no key parameter delete the entire collection */
	key := r.FormValue("key")
	if len(key) == 0 {

		Fuzzy.stores_lock.Lock()
		delete(Fuzzy.stores, parameters["store"])
		Fuzzy.stores_lock.Unlock()

		Fuzzy.stats_lock.Lock()
		delete(Fuzzy.stats, parameters["store"])
		Fuzzy.stats_lock.Unlock()

		return
	}

	deleted := store.Delete(parameters["key"])
	if deleted {
		fmt.Fprintf(w, "Successfully deleted the key")
		IncrementStats(parameters["store"], "/fuzzy DELETE")
	} else {
		http.Error(w, "The key you deleted does not exist", http.StatusBadRequest)
	}
}

func FuzzyHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		GetKeyHandler(w, r)
		return
	case "PUT":
		AddKeyValueHandler(w, r)
		return
	case "DELETE":
		DeleteKeyHandler(w, r)
		return
	case "POST":
		NewStoreHandler(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
