package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getKeyBatchHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store", "keys", "distance"}, w, r)
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

	/* We unmarshall the list of keys */
	var keys []string
	err = json.Unmarshal([]byte(parameters["keys"]), &keys)
	if err != nil {
		parameterError(w, "keys (JSON)")
		return
	}

	result := make([]string, len(keys))

	/* We treat exact matching here */
	if distance == 0 {
		for i, key := range keys {
			value, present := store.Get(key)
			if present {
				result[i] = value
			}
		}
		incrementStats(parameters["store"], "/fuzzy/batch GET")
		jsonResponse, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(jsonResponse))
		return
	}

	/* Approximate matching here */
	results, err := strconv.Atoi(r.FormValue("results"))
	if err != nil {
		parameterError(w, "results")
		return
	}

	c := make(chan struct {
		string
		int
	})

	for i, key := range keys {
		go func(k string, j int) {
			fuzzyResults := store.Query(k, distance, results)
			jsonResponse, _ := json.Marshal(fuzzyResults)
			c <- struct {
				string
				int
			}{string(jsonResponse), j}
		}(key, i)
	}

	for steps := 0; steps < len(keys); steps++ {
		s := <-c
		key, i := s.string, s.int
		result[i] = key
	}

	incrementStats(parameters["store"], "/fuzzy/batch GET")
	jsonResponse, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))
}

func addBatchKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := requireParameters([]string{"store", "dictionary"}, w, r)
	if !valid {
		return
	}

	store, present := getStore(parameters["store"])
	if !present {
		parameterError(w, "store (existent)")
		return
	}

	dict := make(map[string]string)
	err := json.Unmarshal([]byte(parameters["dictionary"]), &dict)
	if err != nil {
		parameterError(w, "dictionary (JSON)")
		return
	}

	for key, value := range dict {
		store.Set(key, value)
	}
	incrementStats(parameters["store"], "/fuzzy/batch PUT")
	fmt.Fprintf(w, "Successfully set the keys")
}

// BatchHandler handles all requests regardin the described API to process
// multiple keys and values at once in order to reduce latency.
func BatchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getKeyBatchHandler(w, r)
		return
	case "PUT":
		addBatchKeyValueHandler(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
