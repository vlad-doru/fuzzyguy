package demo

import (
	"bufio"
	"github.com/vlad-doru/fuzzyguy/server"
	"log"
	"net/http"
	"os"
	"strings"
)

func MonitorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "demo/static/monitor.html")
}

func DemoHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "demo/static/index.html")
}

func EnglishHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("demo/data/english.dat")
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewScanner(file)

	store, _ := server.GetStore("demostore")
	for reader.Scan() {
		split := strings.Split(reader.Text(), "\t")
		store.Set(split[0], split[1])
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}
