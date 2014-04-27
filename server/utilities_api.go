package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	json_response, err := json.Marshal(Fuzzy.stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

	}
	fmt.Fprintf(w, string(json_response))
	w.Header().Set("Content-Type", "application/json")
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
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
