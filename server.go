package main

import (
	"fmt"
	"github.com/vlad-doru/fuzzyguy/levenshtein"
)

func main() {
	fmt.Println(levenshtein.Distance("ana", "ama"))
}
