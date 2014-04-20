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

func GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "key", "distance"}, w, r)
	if !valid {
		return
	}

	store, present := stores[parameters["store"]]
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
	fmt.Fprintf(w, string(json_response))
}

func NewStoreHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	_, present := stores[parameters["store"]]
	if present {
		http.Error(w, "This store already exists", http.StatusBadRequest)
		return
	}

	stores[parameters["store"]] = fuzzy.NewFuzzyService()
	fmt.Fprintf(w, "Store has successfully been created")
	w.WriteHeader(http.StatusCreated)
}

func AddKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "key", "value"}, w, r)
	if !valid {
		return
	}

	store, present := stores[parameters["store"]]
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	store.Set(parameters["key"], parameters["value"])
	fmt.Fprintf(w, "Successfully set the key")
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "key"}, w, r)
	if !valid {
		return
	}

	store, present := stores[parameters["store"]]
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	deleted := store.Delete(parameters["key"])
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
