package internal

type empty struct{}

func ChunkSlice(slice []int, size int) [][]int {
	sliceLen := len(slice)
	if sliceLen == 0 || size <= 0 {
		return [][]int{}
	}

	chunkCnt := sliceLen / size
	if sliceLen%size != 0 {
		chunkCnt += 1
	}

	step := size
	if step > sliceLen {
		step = sliceLen
	}

	result := make([][]int, chunkCnt)
	for i := range result {
		from, to := i*step, i*step+step
		if to > sliceLen {
			to = sliceLen
		}

		result[i] = slice[from:to]
	}

	return result
}

func FilterSlice(slice []int) []int {
	allowedValues := [4]int{2, 4, 6, 8}

	mapOfPredicate := make(map[int]empty, len(allowedValues))
	for _, val := range allowedValues {
		mapOfPredicate[val] = empty{}
	}

	result := make([]int, 0, len(slice))
	for _, val := range slice {
		if _, ok := mapOfPredicate[val]; !ok {
			continue
		}

		result = append(result, val)
	}

	return result
}

func FlipMap(list map[int]int) map[int]int {
	result := make(map[int]int, len(list))
	for index, val := range list {
		result[val] = index
	}

	return result
}
