package levenshtein

const HistogramMod = (1 << 5) - 1

func NewMatrix(dx, dy int) [][]int {
	m := make([][]int, dx)
	for i, _ := range m {
		m[i] = make([]int, dy)
	}
	return m
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Min3(x, y, z int) int {
	if x < y && x < z {
		return x
	}
	if y < z {
		return y
	}
	return z
}

// We make these global so that we avoid reallocation
var v0, v1 = make([]int, 20), make([]int, 20)

func Distance(source, target string) int {
	source_len, target_len := len(source), len(target)

	if len(v0) <= target_len {
		v0, v1 = make([]int, target_len+1), make([]int, target_len+1)
	}

	for i := 0; i <= target_len; i++ {
		v0[i] = i
	}

	cost := 0
	for i := 0; i < source_len; i++ {
		v1[0] = i + 1
		for j := 0; j < target_len; j++ {
			cost = 0
			if source[i] != target[j] {
				cost = 1
			}
			v1[j+1] = Min3(v1[j]+1, v0[j+1]+1, v0[j]+cost)
		}
		v0, v1 = v1, v0
	}

	return v0[target_len]
}

func DistanceThreshold(source, target string, threshold int) (int, bool) {
	source_len, target_len := len(source), len(target)

	if source_len > target_len {
		source, target = target, source
		source_len, target_len = target_len, source_len
	}

	diff := target_len - source_len

	if len(v0) <= target_len {
		v0, v1 = make([]int, target_len+1), make([]int, target_len+1)
	}

	for i := 0; i <= target_len; i++ {
		v0[i] = i
	}

	cost, lower := 0, 0 // In the lower variable we will keep the possible lower bound at each step
	for i := 0; i < source_len; i++ {
		start, stop := Max(0, i-threshold), Min(target_len, i+diff+threshold)
		if start == 0 {
			v1[start] = i + 1
		} else {
			cost = 0
			if source[i-1] != target[start-1] {
				cost = 1
			}
			v1[start] = Min(v0[start]+1, v0[start-1]+cost)
		}
		lower = v1[start]
		for j := start; j < stop-1; j++ {
			cost = 0
			if source[i] != target[j] {
				cost = 1
			}
			v1[j+1] = Min3(v1[j]+1, v0[j+1]+1, v0[j]+cost)
			lower = Min(v1[j+1], lower)
		}
		cost = 0
		if source[i] != target[stop-1] {
			cost = 1
		}
		v1[stop] = Min(v1[stop-1]+1, v0[stop-1]+cost)
		lower = Min(v1[stop], lower)
		// If the lower bound is higher than the threshold we return false
		if lower > threshold {
			return -1, false
		}
		v0, v1 = v1, v0
	}

	return v0[target_len], v0[target_len] <= threshold

}

func ComputeHistogram(s string) uint32 {
	var result uint32 = 0
	for _, c := range s {
		result ^= (1 << (HistogramMod & uint(c)))
	}
	return result
}

/* This function computes the number of different bits in both histogram
   and then adds length_diff to that difference then divides that by 2 */
func LowerBound(histogram_source, histogram_target uint32, length_diff int) int {
	diff := histogram_target ^ histogram_source
	diff -= ((diff >> 1) & 0x55555555)
	diff = (diff & 0x33333333) + ((diff >> 2) & 0x33333333)
	return (int((((diff+(diff>>4))&0x0F0F0F0F)*0x01010101)>>24) + length_diff) >> 1
}

/* Here we compute an extended histogram which allows us to have a better filter
   for fetching a lower bound for the Levenshtein distance. However these functions
   should only be used as a second filter since they are much slower than the pervious
   ones and also take up more memory */

const BUCKET_BITS = 8

var buckets = make([]uint8, 64/BUCKET_BITS)

func ComputeExtendedHistogram(s string) uint64 {
	for i, _ := range buckets {
		buckets[i] = 0
	}
	for _, c := range s {
		index := int(c) % len(buckets)
		if buckets[index] != ((1 << BUCKET_BITS) - 1) {
			buckets[index]++
		}
	}
	var result uint64 = 0
	for i, value := range buckets {
		result += (uint64(value) << uint(i*BUCKET_BITS))
	}
	return result
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ExtendedLowerBound(histogram_source, histogram_target uint64, length_diff int) int {
	bit_mask := (1 << BUCKET_BITS) - 1
	result := length_diff
	for i := 0; i < 64/BUCKET_BITS; i++ {
		source_bucket := int(histogram_source>>uint(i*BUCKET_BITS)) & bit_mask
		target_bucket := int(histogram_target>>uint(i*BUCKET_BITS)) & bit_mask
		result += Abs(target_bucket - source_bucket)
	}
	return result >> 1
}
