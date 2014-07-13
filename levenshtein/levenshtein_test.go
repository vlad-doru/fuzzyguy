package levenshtein

import (
	"testing"
)

func TestDistanceThreshold(t *testing.T) {
	threshold := 2
	var testCases = []struct {
		source   string
		target   string
		distance int
		within   bool
	}{
		{"", "aa", 2, true},
		{"a", "aa", 1, true},
		{"a", "aaa", 2, true},
		{"", "", 0, true},
		{"a", "bcaa", -1, false},
		{"aaa", "aba", 1, true},
		{"aaa", "abcd", -1, false},
		{"a", "a", 0, true},
		{"ab", "aabc", 2, true},
		{"abc", "", -1, false},
		{"aa", "a", 1, true},
		{"aaaaa", "a", -1, false},
		{"informatica", "fmi unibuc", -1, false},
	}
	for _, testCase := range testCases {
		distance, within := DistanceThreshold(testCase.source, testCase.target, threshold)

		if within != testCase.within {
			t.Log("Distance between",
				testCase.source,
				"and",
				testCase.target,
				"computed as",
				within,
				", should be",
				testCase.within)
			t.Error("Failed to compute threshold properly")
		}

		if within && distance != testCase.distance {
			t.Log("Distance between",
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

func BenchmarkLevenshteinThreshold(b *testing.B) {
	source := "informatcia supre"
	target := "informatica super"
	for n := 0; n < b.N; n++ {
		DistanceThreshold(source, target, 3)
	}
}

func BenchmarkLevenshteinThresholdStop(b *testing.B) {
	source := "informaticasapre"
	target := "informatica super"
	for n := 0; n < b.N; n++ {
		DistanceThreshold(source, target, 3)
	}
}

func TestHistogram(t *testing.T) {
	var testCases = []string{"ana", "are", "incredibil", "inexplicabil", "extraveral"}
	for _, s := range testCases {
		aux := make([]int, 32)
		for _, c := range s {
			aux[c%32]++
		}
		var trueValue uint32
		for i := 0; i < 32; i++ {
			if aux[i]%2 == 1 {
				trueValue |= (1 << uint(i))
			}
		}
		if trueValue != ComputeHistogram(s) {
			t.Log("Bad histogram for ", s)
			t.Error("Didn't compute the histogram properly")
		}
	}
}

func BenchmarkHistogram(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ComputeHistogram("informatica fmi unibuc")
	}
}

func TestExtendedHistogram(t *testing.T) {
	var testCases = []string{"ana", "are", "incredibil", "inexplicabil", "extraveral"}
	for _, s := range testCases {
		aux := make([]int, 64/bucketBits)
		for _, c := range s {
			aux[int(c)%(len(aux))]++
		}
		histogram := ComputeExtendedHistogram(s)
		bitMask := uint64(1<<bucketBits) - 1
		for i := 0; i < 64/bucketBits; i++ {
			if aux[i] != int(histogram&bitMask) {
				t.Log("Bad histogram for ", s)
				t.Error("Didn't compute the histogram properly")
				break
			}
			histogram >>= bucketBits
		}
	}
}

func BenchmarkExtendedHistogram(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ComputeExtendedHistogram("informatica fmi unibuc")
	}
}

func TestLowerBound(t *testing.T) {
	var testCases = []struct {
		source   uint32
		target   uint32
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
			t.Log("Difference between",
				testCase.source,
				"and",
				testCase.target,
				"with length_diff 1 computed as",
				distance,
				", should be",
				testCase.distance)
			t.Error("Failed to compute proper Lower Bound")
		}
	}
}

func TestExtendedLowerBound(t *testing.T) {
	var testCases = []struct {
		source   uint64
		target   uint64
		distance int
	}{
		{1, 0, 1},
		{(2 << bucketBits) + 3, (1 << (bucketBits * 2)) + 1, 3},
		{(1 << bucketBits) + 3, (2 << (bucketBits * 2)) + 2, 2},
	}
	for _, testCase := range testCases {
		distance := ExtendedLowerBound(testCase.source, testCase.target, 1)
		if distance != testCase.distance {
			t.Log("Difference between",
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

func BenchmarkExtendedLowerBound(b *testing.B) {
	hist1, hist2 := ComputeExtendedHistogram("informatica"), ComputeExtendedHistogram("fmi unibuc")
	diff := 1 // diferenta intre cele doua siruri
	for n := 0; n < b.N; n++ {
		ExtendedLowerBound(hist1, hist2, diff)
	}
}
