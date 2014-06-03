package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func GetKeyBatchHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "keys", "distance"}, w, r)
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

	/* We unmarshall the list of keys */
	var keys []string
	err = json.Unmarshal([]byte(parameters["keys"]), &keys)
	if err != nil {
		ParameterError(w, "keys (JSON)")
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
		IncrementStats(parameters["store"], "/fuzzy/batch GET")
		json_response, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(json_response))
		return
	}

	/* Approximate matching here */
	results, err := strconv.Atoi(r.FormValue("results"))
	if err != nil {
		ParameterError(w, "results")
		return
	}

	c := make(chan struct {
		string
		int
	})

	for i, key := range keys {
		go func(k string, j int) {
			fuzzy_results := store.Query(k, distance, results)
			json_response, _ := json.Marshal(fuzzy_results)
			c <- struct {
				string
				int
			}{string(json_response), j}
		}(key, i)
	}

	for steps := 0; steps < len(keys); steps++ {
		s := <-c
		key, i := s.string, s.int
		result[i] = key
	}

	IncrementStats(parameters["store"], "/fuzzy/batch GET")
	json_response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json_response))
}

func AddBatchKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "dictionary"}, w, r)
	if !valid {
		return
	}

	store, present := GetStore(parameters["store"])
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	dict := make(map[string]string)
	err := json.Unmarshal([]byte(parameters["dictionary"]), &dict)
	if err != nil {
		ParameterError(w, "dictionary (JSON)")
		return
	}

	for key, value := range dict {
		store.Set(key, value)
	}
	IncrementStats(parameters["store"], "/fuzzy/batch PUT")
	fmt.Fprintf(w, "Successfully set the keys")
}

func BatchHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		GetKeyBatchHandler(w, r)
		return
	case "PUT":
		AddBatchKeyValueHandler(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
