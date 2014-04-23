package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	w.Header().Set("Content-Type", "application/json")
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

	fmt.Fprintf(w, "Successfully set the keys")
}

func BatchHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
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

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/admin.html")
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

	store, _ := stores["english"]
	for reader.Scan() {
		split := strings.Split(reader.Text(), "\t")
		store.Set(split[0], split[1])
	}
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
	http.HandleFunc("/static/", StaticHandler)
	http.HandleFunc("/demo/", StaticHandler)

	// Admin handler
	http.HandleFunc("/admin", AdminHandler)

	// Demo handler
	http.HandleFunc("/demo", DemoHandler)
	// Demo load english dictionary
	http.HandleFunc("/demo/loadenglish", EnglishHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", configuration.Port), nil)
}
