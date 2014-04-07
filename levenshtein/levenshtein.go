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
	if len(source) == 0 {
		return len(target)
	}
	if len(target) == 0 {
		return len(source)
	}

	dist := NewMatrix(len(source) + 1, len(target) + 1)
	
	for i := 0; i<=len(source); i++ {
		dist[i][0] = i
	}
	for i := 0; i<=len(target); i++ {
		dist[0][i] = i
	}

	for i := 1; i <= len(source); i++ {
		for j := 1; j <= len(target); j++ {
			cost := 0
			if source[i - 1] != target[j - 1] {
				cost = 1
			}
			dist[i][j] = Min(dist[i-1][j] + 1, 
							 dist[i][j-1] + 1,
							 dist[i-1][j-1] + cost)
		}
	}

	return dist[len(source)][len(target)]
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
    return (int((((diff + (diff >> 4)) & 0x0F0F0F0F) * 0x01010101) >> 24) + length_diff) >> 1
}