package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
		application_stats.mutex.Lock()
		application_stats.Stores[parameters["store"]].Queries["GET Excat"] += 1
		application_stats.mutex.Unlock()
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
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json_response))
	application_stats.mutex.Lock()
	application_stats.Stores[parameters["store"]].Queries["GET Fuzzy"] += 1
	application_stats.mutex.Unlock()
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
	application_stats.Stores[parameters["store"]] = StoreStatistics{map[string]int{"POST": 1}}
	fmt.Fprintf(w, "Store has successfully been created")
	w.WriteHeader(http.StatusCreated)
	application_stats.mutex.Lock()
	application_stats.Stores[parameters["store"]].Queries["POST NewStore"] += 1
	application_stats.mutex.Unlock()
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
	application_stats.mutex.Lock()
	application_stats.Stores[parameters["store"]].Queries["PUT AddKey"] += 1
	application_stats.mutex.Unlock()
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store"}, w, r)
	if !valid {
		return
	}

	store, present := stores[parameters["store"]]
	if !present {
		ParameterError(w, "store (existent)")
		return
	}

	/* If there is no key parameter delete the entire collection */
	key := r.FormValue("key")
	if len(key) == 0 {
		delete(stores, parameters["store"])
		application_stats.mutex.Lock()
		delete(application_stats.Stores, parameters["store"])
		application_stats.mutex.Unlock()
		return
	}

	deleted := store.Delete(parameters["key"])
	if deleted {
		fmt.Fprintf(w, "Successfully deleted the key")
		application_stats.mutex.Lock()
		application_stats.Stores[parameters["store"]].Queries["DELETE Key"] += 1
		application_stats.mutex.Unlock()
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

func GetKeyBatchHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "keys", "distance"}, w, r)
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
		application_stats.mutex.Lock()
		application_stats.Stores[parameters["store"]].Queries["GET Batch"] += 1
		application_stats.mutex.Unlock()
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
		go func(key string, i int) {
			fuzzy_results := store.Query(key, distance, results)
			json_response, _ := json.Marshal(fuzzy_results)
			c <- struct {
				string
				int
			}{string(json_response), i}
		}(key, i)
	}

	for steps := 0; steps < len(keys); steps++ {
		s := <-c
		key, i := s.string, s.int
		result[i] = key
	}

	json_response, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json_response))
	application_stats.mutex.Lock()
	application_stats.Stores[parameters["store"]].Queries["GET Batch"] += 1
	application_stats.mutex.Unlock()
}

func AddBatchKeyValueHandler(w http.ResponseWriter, r *http.Request) {
	parameters, valid := RequireParameters([]string{"store", "dictionary"}, w, r)
	if !valid {
		return
	}

	store, present := stores[parameters["store"]]
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
	application_stats.mutex.Lock()
	application_stats.Stores[parameters["store"]].Queries["PUT Batch"] += 1
	application_stats.mutex.Unlock()
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

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "static/admin.html" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	} else {
		http.ServeFile(w, r, r.URL.Path[1:])
	}
}

func MonitorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "demo/monitor.html")
}

func DemoHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "demo/index.html")
}

func EnglishHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("demo/data/english.dat")
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewScanner(file)

	store, _ := stores["demostore"]
	for reader.Scan() {
		split := strings.Split(reader.Text(), "\t")
		store.Set(split[0], split[1])
	}
}

type StoreStatistics struct {
	Queries map[string]int
}

type Statistics struct {
	Stores map[string]StoreStatistics
	mutex  *sync.RWMutex
}

var application_stats = Statistics{make(map[string]StoreStatistics), new(sync.RWMutex)}

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	json_response, err := json.Marshal(application_stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

	}
	fmt.Fprintf(w, string(json_response))
	w.Header().Set("Content-Type", "application/json")

}

func DemoTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	parameters, valid := RequireParameters([]string{"testsize", "distance", "results"}, w, r)
	if !valid {
		return
	}
	cmd := exec.Command("python", "test/test.py", parameters["testsize"], parameters["distance"], parameters["results"])
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, out.String())
	w.Header().Set("Content-Type", "application/json")

}

type Configuration struct {
	Port     string
	Admin    string
	Password string
}

func LoadConfiguration() *Configuration {
	// We always get our configuration from conf.json
	file, _ := os.Open("conf.json")

	decoder := json.NewDecoder(file)
	configuration := new(Configuration)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("[FATAL ERROR]", err)
		return nil
	}
	return configuration
}

func main() {

	configuration := LoadConfiguration()
	if configuration == nil {
		return
	}

	// API Handlers
	http.HandleFunc("/fuzzy", FuzzyHandler)
	http.HandleFunc("/fuzzy/batch", BatchHandler)

	// Serve all static files
	http.HandleFunc("/demo/css/", StaticHandler)
	http.HandleFunc("/demo/img/", StaticHandler)
	http.HandleFunc("/demo/js/", StaticHandler)

	// Demo handler
	http.HandleFunc("/demo", DemoHandler)
	// Demo Monitor handler
	http.HandleFunc("/demo/monitor", MonitorHandler)
	// Demo load english dictionary
	http.HandleFunc("/demo/loadenglish", EnglishHandler)

	// Statistics handler for our service
	http.HandleFunc("/stats", StatsHandler)

	// Test handler for our service
	http.HandleFunc("/demo/test", DemoTestHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", configuration.Port), nil)
}
