package internal

import (
	"fmt"

	"ova-method-api/internal/model"
)

type empty struct{}

var (
	DuplicateKeyErr     = fmt.Errorf("duplicate key")
	InvalidChunkSizeErr = fmt.Errorf("invalid chunk size")
)

func ChunkSlice(slice []int, size int) ([][]int, error) {
	sliceLen := len(slice)
	chunkCnt, step, err := calcChunkParams(sliceLen, size)
	if err == InvalidChunkSizeErr {
		return [][]int{}, InvalidChunkSizeErr
	}

	result := make([][]int, chunkCnt)
	for i := range result {
		from, to := calcSliceRangeForChunk(i, step, sliceLen)
		result[i] = slice[from:to]
	}

	return result, nil
}

func calcChunkParams(sliceLen, size int) (chunkCnt int, step int, err error) {
	if size <= 0 {
		return 0, 0, InvalidChunkSizeErr
	}
	if sliceLen == 0 {
		return 0, 0, nil
	}

	chunkCnt = sliceLen / size
	if sliceLen%size != 0 {
		chunkCnt += 1
	}

	step = size
	if step > sliceLen {
		step = sliceLen
	}

	return chunkCnt, step, nil
}

func calcSliceRangeForChunk(chunk, size, sliceLen int) (from int, to int) {
	from, to = chunk*size, chunk*size+size
	if to > sliceLen {
		to = sliceLen
	}
	return from, to
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
		if _, ok := result[val]; ok {
			panic(DuplicateKeyErr)
		}
		result[val] = index
	}

	return result
}

func ListOfMethodToUserMap(list []model.Method) (map[uint64]model.Method, error) {
	result := make(map[uint64]model.Method, len(list))
	for _, method := range list {
		if _, ok := result[method.UserId]; ok {
			return nil, DuplicateKeyErr
		}
		result[method.UserId] = method
	}

	return result, nil
}

func ListOfMethodToChunkSlice(list []model.Method, size int) ([][]model.Method, error) {
	sliceLen := len(list)
	chunkCnt, step, err := calcChunkParams(sliceLen, size)
	if err == InvalidChunkSizeErr {
		return [][]model.Method{}, InvalidChunkSizeErr
	}

	result := make([][]model.Method, chunkCnt)
	for i := range result {
		from, to := calcSliceRangeForChunk(i, step, sliceLen)
		result[i] = list[from:to]
	}

	return result, nil
}
