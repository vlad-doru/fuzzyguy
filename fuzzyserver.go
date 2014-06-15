package main

import (
	"encoding/json"
	"fmt"
	"./server"
	"log"
	"net/http"
	"os"
	"runtime"
)

type configuration struct {
	Port string
}

func loadConfiguration() *configuration {
	// We always get our conf from conf.json
	file, _ := os.Open("conf.json")

	decoder := json.NewDecoder(file)
	conf := new(configuration)
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("[FATAL ERROR]", err)
		return nil
	}
	return conf
}

func main() {

	conf := loadConfiguration()
	if conf == nil {
		return
	}
	// We set the maximum number of cores to be used
	runtime.GOMAXPROCS(runtime.NumCPU())

	// API Handlers
	http.HandleFunc("/fuzzy", server.FuzzyHandler)
	http.HandleFunc("/fuzzy/batch", server.BatchHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil))
}
