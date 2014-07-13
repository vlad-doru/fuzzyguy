package levenshtein

// Number of bits for histogram
const histogramMod = (1 << 5) - 1

func newMatrix(dx, dy int) [][]int {
	m := make([][]int, dx)
	for i := range m {
		m[i] = make([]int, dy)
	}
	return m
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func min3(x, y, z int) int {
	if x < y && x < z {
		return x
	}
	if y < z {
		return y
	}
	return z
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DistanceThreshold computes the Levenshtein distance between two strings
// if and only if it is smaller than a specific threshold
//
// Arugments:
// source, target (string): thw two strings to compute the distance for
// threshold (int): the threshold of the Levenshtein distance
//
// Returns: (int, bool) Levenshtein distance and if it is lower than the
// threshold. The first value is valid iff the second one is true.
func DistanceThreshold(source, target string, threshold int) (int, bool) {
	sourceLen := len(source)
	targetLen := len(target)

	if sourceLen > targetLen {
		source, target = target, source
		sourceLen, targetLen = targetLen, sourceLen
	}

	diff := targetLen - sourceLen

	v0, v1 := make([]int, targetLen+1), make([]int, targetLen+1)

	for i := 0; i <= targetLen; i++ {
		v0[i] = i
	}

	cost, lower := 0, 0 // Lower bound at each step
	for i := 0; i < sourceLen; i++ {
		start, stop := max(0, i-threshold), min(targetLen, i+diff+threshold)
		if start == 0 {
			v1[start] = i + 1
		} else {
			cost = 0
			if source[i-1] != target[start-1] {
				cost = 1
			}
			v1[start] = min(v0[start]+1, v0[start-1]+cost)
		}
		lower = v1[start]
		for j := start; j < stop-1; j++ {
			cost = 0
			if source[i] != target[j] {
				cost = 1
			}
			v1[j+1] = min3(v1[j]+1, v0[j+1]+1, v0[j]+cost)
			lower = min(v1[j+1], lower)
		}
		cost = 0
		if source[i] != target[stop-1] {
			cost = 1
		}
		v1[stop] = min(v1[stop-1]+1, v0[stop-1]+cost)
		lower = min(v1[stop], lower)
		// If the lower bound is higher than the threshold we return false
		if lower > threshold {
			return -1, false
		}
		v0, v1 = v1, v0
	}

	return v0[targetLen], v0[targetLen] <= threshold

}

// ComputeHistogram calculates the 32bit histogram for a specific string
//
// Arugments:
// s (string): The string for which we want to calculate the histrogram
//
// Returns: (uint32) the computed 32bit histogram
func ComputeHistogram(s string) uint32 {
	var result uint32
	for _, c := range s {
		result ^= (1 << (histogramMod & uint(c)))
	}
	return result
}

// LowerBound computes the number of different bits in both histograms
// and then adds lengthDiff to that difference then divides that by 2
//
// Arguments:
// histogramSource, histogramTarget (uint32): the two histograms
// lengthDiff (int): the difference of length between the two corresponding
// strings
func LowerBound(histogramSource, histogramTarget uint32, lengthDiff int) int {
	diff := histogramTarget ^ histogramSource
	diff = diff - ((diff >> 1) & 0x55555555)
	diff = (diff & 0x33333333) + ((diff >> 2) & 0x33333333)
	diff = (((diff + (diff >> 4)) & 0x0F0F0F0F) * 0x01010101) >> 24
	return ((int)(diff) + lengthDiff) >> 1
}

// Here we compute an extended histogram which allows us to have a better filter
// for fetching a lower bound for the Levenshtein distance. However these functions
// should only be used as a second filter since they are much slower than the pervious
// ones and also take up more memory

const bucketBits = 2
const bitMask = (1 << bucketBits) - 1

// ComputeExtendedHistogram takes a string and outputs a more refined, 64bit
// histogram which represents a refined version of the 32bit one
//
// Arguments:
// s (string): the string which we want to compute the histogram for
//
// Returns: (uint64) the 64bit computed histogram
func ComputeExtendedHistogram(s string) uint64 {
	buckets := make([]uint8, 64/bucketBits)
	for i := range buckets {
		buckets[i] = 0
	}
	for _, c := range s {
		index := int(c) % len(buckets)
		if buckets[index] != bitMask {
			buckets[index]++
		}
	}
	var result uint64
	for i, value := range buckets {
		result += (uint64(value) << uint(i*bucketBits))
	}
	return result
}

// ExtendedLowerBound takes two 64bit histograms and the corresponding string
// length difference and outputs a lower bound for the Levenshtein distance
//
// Arugments:
// histogramSource, histogramTarget (uint64): the two histograms
// lengthDiff (int): the corresponding string length difference
//
// Returns: (int) the lower bound for the Levenshtein distance based on these
// histograms.
func ExtendedLowerBound(histogramSource, histogramTarget uint64, lengthDiff int) int {
	result := lengthDiff
	for i := 0; i < 64/bucketBits; i++ {
		sourceBucket := int(histogramSource) & bitMask
		targetBucket := int(histogramTarget) & bitMask
		result += abs(targetBucket - sourceBucket)
		histogramSource >>= bucketBits
		histogramTarget >>= bucketBits
	}
	return result >> 1
}
