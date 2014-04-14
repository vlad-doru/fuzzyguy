package main

import (
	"fmt"
	"github.com/vlad-doru/fuzzyguy/fuzzy"
	"net/http"
)

var stores = map[string]*

func FuzzyHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "Hi there, GET!")
	case "POST":
		fmt.Fprintf(w, "Hi there, POST!")
	case "PUT":
		fmt.Fprintf(w, "Hi there, PUT!")
	case "DELETE":
		fmt.Fprintf(w, "Hi there, DELETE!")
	default:
		http.Error(w, http.StatusText(405), 405)
	}
}

func main() {
	http.HandleFunc("/fuzzy", FuzzyHandler)
	http.ListenAndServe(":8080", nil)
}
