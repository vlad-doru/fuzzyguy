package main

import (
	"encoding/json"
	"fmt"
	"github.com/vlad-doru/fuzzyguy/demo"
	"github.com/vlad-doru/fuzzyguy/server"
	"net/http"
	"os"
	"runtime"
)

type Configuration struct {
	Port string
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
	// We set the maximum number of cores to be used
	runtime.GOMAXPROCS(runtime.NumCPU())

	// API Handlers
	http.HandleFunc("/fuzzy", server.FuzzyHandler)
	http.HandleFunc("/fuzzy/batch", server.BatchHandler)
	// Statistics handler for our service
	http.HandleFunc("/fuzzy/stats", server.StatsHandler)
	// Test handler for our service
	http.HandleFunc("/fuzzy/test", server.TestHandler)

	// Demo Handlers
	http.HandleFunc("/demo/static/", demo.StaticHandler)
	http.HandleFunc("/demo", demo.DemoHandler)
	http.HandleFunc("/demo/monitor", demo.MonitorHandler)
	http.HandleFunc("/demo/loadenglish", demo.EnglishHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", configuration.Port), nil)
}
