package main

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"net/http"
	"strconv"
)

var stores map[string]*fuzzy.FuzzyService = make(map[string]*fuzzy.FuzzyService)

func ParameterError(w http.ResponseWriter, parameter string) {
	http.Error(w, fmt.Sprintf("Please provide a valid %s", parameter), http.StatusBadRequest)
}

func GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	storeName := r.FormValue("store")
	if len(storeName) == 0 {
		ParameterError(w, "store")
		return
	}
	store, present := stores[storeName]
	if !present {
		ParameterError(w, "store")
		return
	}

	key := r.FormValue("key")
	if len(key) == 0 {
		ParameterError(w, "key")
		return
	}

	/* We always require a distance parameter in order to make every request more explicit
	   about whether we would like to perform and exact match or an approximate one */

	distance, err := strconv.Atoi(r.FormValue("distance"))
	if err != nil {
		ParameterError(w, "distance")
		return
	}

	/* We treat exact matching here */
	if distance == 0 {
		value, present := store.Get(key)
		if !present {
			ParameterError(w, "key (Existent)")
			return
		}
		fmt.Fprintf(w, value)
		return
	}

	/* Approximate matching here */
	results, err := strconv.Atoi(r.FormValue("results"))
	if err != nil {
		http.Error(w, "Please provide a results parameter", http.StatusBadRequest)
		return
	}
	fuzzy_results := store.Query(key, distance, results)
	json_response, _ := json.Marshal(fuzzy_results)
	fmt.Fprintf(w, string(json_response))
}

func NewStoreHandler(w http.ResponseWriter, r *http.Request) {
	storeName := r.FormValue("store")
	if len(storeName) == 0 {
		ParameterError(w, "store")
		return
	}
	_, present := stores[storeName]
	if present {
		http.Error(w, "This store already exists", http.StatusBadRequest)
		return
	}

	stores[storeName] = fuzzy.NewFuzzyService()
	fmt.Fprintf(w, "Store has successfully been created")
	w.WriteHeader(http.StatusCreated)
}

func AddKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	storeName := r.FormValue("store")
	if len(storeName) == 0 {
		ParameterError(w, "store")
		return
	}
	store, present := stores[storeName]
	if !present {
		ParameterError(w, "store")
		return
	}

	key := r.FormValue("key")
	if len(key) == 0 {
		ParameterError(w, "key")
		return
	}

	value := r.FormValue("value")
	if len(value) == 0 {
		ParameterError(w, "value")
		return
	}

	store.Set(key, value)
	fmt.Fprintf(w, "Successfully set the key")
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	storeName := r.FormValue("store")
	if len(storeName) == 0 {
		ParameterError(w, "store")
		return
	}
	store, present := stores[storeName]
	if !present {
		ParameterError(w, "store")
		return
	}

	key := r.FormValue("key")
	if len(key) == 0 {
		ParameterError(w, "key")
		return
	}

	deleted := store.Delete(key)
	if deleted {
		fmt.Fprintf(w, "Successfully deleted the key")
	} else {
		http.Error(w, "The key you deleted does not exist", http.StatusBadRequest)
	}
}

func FuzzyHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		GetKeyHandler(w, r)
		break
	case "PUT":
		AddKeyValueHandler(w, r)
		break
	case "DELETE":
		DeleteKeyHandler(w, r)
		break
	case "POST":
		NewStoreHandler(w, r)
		break
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/fuzzy", FuzzyHandler)
	http.ListenAndServe(":8080", nil)
}
