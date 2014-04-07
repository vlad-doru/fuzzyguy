package levenshtein

import (
	"testing"
	)

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
	for _, testCase := range testCases {
		distance := Distance(testCase.source, testCase.target)
		if distance != testCase.distance {
			t.Log( "Distance between",
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

func TestHistogram(t *testing.T) {
	var testCases = []string {"ana", "are", "incredibil", "inexplicabil", "extraveral"}
	for _, s := range testCases {
		aux := make([]int, 32)
		for _, c := range s {
			aux[c % 32]++
		}
		var true_value uint32 = 0
		for i := 0; i < 32; i++ {
			if aux[i]%2 == 1 {
				true_value |= (1 << uint(i))
			}
		}
		if true_value != ComputeHistogram(s) {
			t.Log( "Bad histogram for ", s)
			t.Error("Didn't compute the histogram properly")
		}
	}
}

func BenchmarkHistogram(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ComputeHistogram("informatica fmi unibuc")
	}
}

func TestLowerBound(t *testing.T) {
	var testCases = []struct {
		source uint32
		target uint32
		distance int
	}{
		{1, 0, 1},
		{2, 3, 1},
		{4, 4, 0},
		{123, 123, 0},
		{4, 2, 1},
		{4, 3, 2},
	}
	for _, testCase := range testCases {
		distance := LowerBound(testCase.source, testCase.target, 1)
		if distance != testCase.distance {
			t.Log( "Difference between",
				    testCase.source,
					"and",
					testCase.target,
					"computed as",
					distance,
					", should be",
					testCase.distance)
			t.Error("Failed to compute proper Lower Bound")
		}
	}
}

func BenchmarkLowerBound(b *testing.B) {
	hist1, hist2 := ComputeHistogram("informatica"), ComputeHistogram("fmi unibuc")
	diff := 1 // diferenta intre cele doua siruri
	for n := 0; n < b.N; n++ {
		LowerBound(hist1, hist2, diff)
	}
}