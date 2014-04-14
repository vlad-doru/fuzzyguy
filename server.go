package main

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"net/http"
	"strconv"
)

var stores map[string]*fuzzy.FuzzyService = make(map[string]*fuzzy.FuzzyService)

func FuzzyHandler(w http.ResponseWriter, r *http.Request) {
	storeName := r.FormValue("store")
	if len(storeName) == 0 {
		http.Error(w, "Please provide the store name", http.StatusBadRequest)
		return
	}
	store, present := stores[storeName]
	if !present {
		if r.Method != "POST" {
			http.Error(w, "Please provide an existent store name", http.StatusBadRequest)
			return
		}
		stores[storeName] = fuzzy.NewFuzzyService()
		fmt.Fprintf(w, "Store has successfully been created")
		w.WriteHeader(http.StatusCreated)
		return
	}

	if r.Method == "POST" {
		http.Error(w, "This store already exists", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")
	if len(key) == 0 {
		http.Error(w, "Please provide a key parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		distance, err := strconv.Atoi(r.FormValue("distance"))
		if err != nil {
			http.Error(w, "Please provide a distance parameter", http.StatusBadRequest)
			return
		}
		if distance == 0 {
			value, present := store.Get(key)
			if !present {
				http.Error(w, "Key does not exist", http.StatusBadRequest)
				return
			}
			fmt.Fprintf(w, value)
			return
		}
		results, err := strconv.Atoi(r.FormValue("results"))
		if err != nil {
			http.Error(w, "Please provide a results parameter", http.StatusBadRequest)
			return
		}
		fuzzy_results := store.Query(key, distance, results)
		json_response, _ := json.Marshal(fuzzy_results)
		fmt.Fprintf(w, string(json_response))

	case "PUT":
		value := r.FormValue("value")
		if len(value) == 0 {
			http.Error(w, "Please provide a value parameter", http.StatusBadRequest)
			return
		}
		store.Set(key, value)
		fmt.Fprintf(w, "Successfully set the key")
	case "DELETE":
		deleted := store.Delete(key)
		if deleted {
			fmt.Fprintf(w, "Successfully deleted the key")
		} else {
			http.Error(w, "The key you deleted does not exist", http.StatusBadRequest)
		}
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/fuzzy", FuzzyHandler)
	http.ListenAndServe(":8080", nil)
}
