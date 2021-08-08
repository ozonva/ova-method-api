package internal

import (
	"reflect"
	"testing"
)

func TestFlipMap(t *testing.T) {
	testCases := []struct {
		sequence    map[int]int
		expectedSeq map[int]int
		expectedRes map[int]int
	}{
		{
			sequence:    map[int]int{},
			expectedSeq: map[int]int{},
			expectedRes: map[int]int{},
		},
		{
			sequence:    map[int]int{0: 0},
			expectedSeq: map[int]int{0: 0},
			expectedRes: map[int]int{0: 0},
		},
		{
			sequence:    map[int]int{1: 2},
			expectedSeq: map[int]int{1: 2},
			expectedRes: map[int]int{2: 1},
		},
		{
			sequence:    map[int]int{1: 2, 2: 2},
			expectedSeq: map[int]int{1: 2, 2: 2},
			expectedRes: map[int]int{2: 2},
		},
		{
			sequence:    map[int]int{1: 2, 2: 3},
			expectedSeq: map[int]int{1: 2, 2: 3},
			expectedRes: map[int]int{2: 1, 3: 2},
		},
	}

	for index, testCase := range testCases {
		result := FlipMap(testCase.sequence)
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			mutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func TestFilterSlice(t *testing.T) {
	testCases := []struct {
		sequence    []int
		expectedSeq []int
		expectedRes []int
	}{
		{
			sequence:    []int{},
			expectedSeq: []int{},
			expectedRes: []int{},
		},
		{
			sequence:    []int{3},
			expectedSeq: []int{3},
			expectedRes: []int{},
		},
		{
			sequence:    []int{2, 3, 4},
			expectedSeq: []int{2, 3, 4},
			expectedRes: []int{2, 4},
		},
		{
			sequence:    []int{1, 2, 3, 6, 6, 7},
			expectedSeq: []int{1, 2, 3, 6, 6, 7},
			expectedRes: []int{2, 6, 6},
		},
	}

	for index, testCase := range testCases {
		result := FilterSlice(testCase.sequence)
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			mutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func TestChunkSlice(t *testing.T) {
	testCases := []struct {
		chunk       int
		sequence    []int
		expectedSeq []int
		expectedRes [][]int
	}{
		{
			chunk:       0,
			sequence:    []int{},
			expectedSeq: []int{},
			expectedRes: [][]int{},
		},
		{
			chunk:       -1,
			sequence:    []int{},
			expectedSeq: []int{},
			expectedRes: [][]int{},
		},
		{
			chunk:       10,
			sequence:    []int{1},
			expectedSeq: []int{1},
			expectedRes: [][]int{{1}},
		},
		{
			chunk:       2,
			sequence:    []int{1, 2},
			expectedSeq: []int{1, 2},
			expectedRes: [][]int{{1, 2}},
		},
		{
			chunk:       2,
			sequence:    []int{1, 2, 3},
			expectedSeq: []int{1, 2, 3},
			expectedRes: [][]int{{1, 2}, {3}},
		},
	}

	for index, testCase := range testCases {
		result := ChunkSlice(testCase.sequence, testCase.chunk)
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			mutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func testError(t *testing.T, index int, expectedRes, result interface{}) {
	t.Errorf("failed testCase[%d], expected %v got %v", index, expectedRes, result)
}

func mutateError(t *testing.T, index int, expectedRes, result interface{}) {
	t.Errorf(
		"failed testCase[%d], data has been mutated, before %v after %v",
		index, expectedRes, result,
	)
}
