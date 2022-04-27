package chomp

// int8 abs
func abs8(a int8) int8 {
	if a < 0 {
		return -a
	}

	return a
}

func minmax(a, b int8) (int8, int8) {
	if a > b {
		return b, a
	}

	return a, b
}
