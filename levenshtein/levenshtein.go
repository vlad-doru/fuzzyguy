package levenshtein

const HistogramMod = (1 << 5) - 1

func NewMatrix(dx, dy int) [][]int {
	m := make([][]int, dx)
	for i, _ := range m {
		m[i] = make([]int, dy)
	}
	return m
}

func Min(x, y, z int) int {
	if x < y && x < z {
		return x
	}
	if y < z {
		return y
	}
	return z
}

func Distance(source, target string) int {
	source_len, target_len := len(source), len(target)
	if source_len == 0 {
		return target_len
	}
	if target_len == 0 {
		return source_len
	}

	v0, v1 := make([]int, target_len+1), make([]int, target_len+1)

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
			v1[j+1] = Min(v1[j]+1, v0[j+1]+1, v0[j]+cost)
		}
		v0, v1 = v1, v0
	}

	return v0[target_len]
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
