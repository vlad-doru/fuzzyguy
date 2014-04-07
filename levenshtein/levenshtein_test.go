package levenshtein

import (
	"testing"
	)

var testCases = []struct {
	source string
	target string
	distance int
}{
	{"", "a", 1},
	{"a", "aa", 1},
	{"a", "aaa", 2},
	{"", "", 0},
	{"a", "b", 1},
	{"aaa", "aba", 1},
	{"aaa", "ab", 2},
	{"a", "a", 0},
	{"ab", "ab", 0},
	{"a", "", 1},
	{"aa", "a", 1},
	{"aaa", "a", 2},
	{"informatica", "fmi unibuc", 10},
}


func TestNewMatrix(t *testing.T) {
	m := NewMatrix(3, 4) 
	for i := 0; i<3; i++ {
		for j :=0; j<4; j++ {
			if m[i][j] != 0 {
				t.Error("Matrix was not initialized as it was supposed to")
			}
		}
	}
}

func TestLevenshtein(t *testing.T) {
	for _, testCase := range testCases {
		distance := Distance(testCase.source, testCase.target)
		if distance != testCase.distance {
			t.Log(
					"Distance between",
					testCase.source,
					"and",
					testCase.target,
					"computed as",
					distance,
					", should be",
					testCase.distance)
			t.Error("Failed to compute proper Levenshtein Distance")
		}
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Distance("informatica", "fmi unibuc")
	}
}