package levenshtein

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

func LevenshteinDistance(source, target string) int {
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