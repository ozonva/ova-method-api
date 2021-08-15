package internal

import (
	"reflect"
	"testing"

	"ova-method-api/internal/model"
)

func TestFlipMap(t *testing.T) {
	testCases := []struct {
		sequence      map[int]int
		expectedSeq   map[int]int
		expectedRes   map[int]int
		panicExpected bool
	}{
		{
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: map[int]int{},
		},
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
			sequence:      map[int]int{1: 2, 2: 2},
			expectedSeq:   map[int]int{1: 2, 2: 2},
			expectedRes:   map[int]int{2: 2},
			panicExpected: true,
		},
		{
			sequence:    map[int]int{1: 2, 2: 3},
			expectedSeq: map[int]int{1: 2, 2: 3},
			expectedRes: map[int]int{2: 1, 3: 2},
		},
	}

	for index, testCase := range testCases {
		if testCase.panicExpected {
			assertPanic(t, index, func() { FlipMap(testCase.sequence) })
			continue
		}

		result := FlipMap(testCase.sequence)
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testAssertError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			testMutateError(t, index, testCase.expectedRes, result)
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
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: []int{},
		},
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
			testAssertError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			testMutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func TestChunkSlice(t *testing.T) {
	testCases := []struct {
		chunk       int
		sequence    []int
		expectedSeq []int
		expectedRes [][]int
		expectedErr error
	}{
		{
			chunk:       0,
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: [][]int{},
		},
		{
			chunk:       2,
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: [][]int{},
		},
		{
			chunk:       0,
			sequence:    []int{},
			expectedSeq: []int{},
			expectedRes: [][]int{},
		},
		{
			chunk:       -1,
			sequence:    []int{1},
			expectedSeq: []int{1},
			expectedRes: [][]int{},
			expectedErr: InvalidChunkSizeErr,
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
		result, err := ChunkSlice(testCase.sequence, testCase.chunk)
		if err != testCase.expectedErr {
			testWantError(t, index, testCase.expectedErr, err)
		}
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testAssertError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			testMutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func TestListOfMethodToUserMap(t *testing.T) {
	testCases := []struct {
		sequence    []model.Method
		expectedSeq []model.Method
		expectedRes map[uint64]model.Method
		expectedErr error
	}{
		{
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: map[uint64]model.Method{},
		},
		{
			sequence:    []model.Method{},
			expectedSeq: []model.Method{},
			expectedRes: map[uint64]model.Method{},
		},
		{
			sequence:    []model.Method{{UserId: 1}, {UserId: 2}},
			expectedSeq: []model.Method{{UserId: 1}, {UserId: 2}},
			expectedRes: map[uint64]model.Method{1: {UserId: 1}, 2: {UserId: 2}},
		},
		{
			sequence:    []model.Method{{UserId: 1, Value: "1"}, {UserId: 1, Value: "2"}},
			expectedSeq: []model.Method{{UserId: 1, Value: "1"}, {UserId: 1, Value: "2"}},
			expectedRes: map[uint64]model.Method{1: {UserId: 1, Value: "2"}},
			expectedErr: DuplicateKeyErr,
		},
	}

	for index, testCase := range testCases {
		result, err := ListOfMethodToUserMap(testCase.sequence)
		if err != testCase.expectedErr {
			testWantError(t, index, testCase.expectedErr, err)
		}
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testAssertError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			testMutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func TestListOfMethodToChunkSlice(t *testing.T) {
	testCases := []struct {
		chunk       int
		sequence    []model.Method
		expectedSeq []model.Method
		expectedRes [][]model.Method
		expectedErr error
	}{
		{
			chunk:       0,
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: [][]model.Method{},
		},
		{
			chunk:       2,
			sequence:    nil,
			expectedSeq: nil,
			expectedRes: [][]model.Method{},
		},
		{
			chunk:       0,
			sequence:    []model.Method{},
			expectedSeq: []model.Method{},
			expectedRes: [][]model.Method{},
		},
		{
			chunk:       -1,
			sequence:    []model.Method{{Value: "1"}},
			expectedSeq: []model.Method{{Value: "1"}},
			expectedRes: [][]model.Method{},
			expectedErr: InvalidChunkSizeErr,
		},
		{
			chunk:       10,
			sequence:    []model.Method{{Value: "1"}},
			expectedSeq: []model.Method{{Value: "1"}},
			expectedRes: [][]model.Method{{{Value: "1"}}},
		},
		{
			chunk:       2,
			sequence:    []model.Method{{Value: "1"}, {Value: "2"}},
			expectedSeq: []model.Method{{Value: "1"}, {Value: "2"}},
			expectedRes: [][]model.Method{{{Value: "1"}, {Value: "2"}}},
		},
		{
			chunk:       2,
			sequence:    []model.Method{{Value: "1"}, {Value: "2"}, {Value: "3"}},
			expectedSeq: []model.Method{{Value: "1"}, {Value: "2"}, {Value: "3"}},
			expectedRes: [][]model.Method{{{Value: "1"}, {Value: "2"}}, {{Value: "3"}}},
		},
	}

	for index, testCase := range testCases {
		result, err := ListOfMethodToChunkSlice(testCase.sequence, testCase.chunk)
		if err != testCase.expectedErr {
			testWantError(t, index, testCase.expectedErr, err)
		}
		if !reflect.DeepEqual(result, testCase.expectedRes) {
			testAssertError(t, index, testCase.expectedRes, result)
		}
		if !reflect.DeepEqual(testCase.sequence, testCase.expectedSeq) {
			testMutateError(t, index, testCase.expectedRes, result)
		}
	}
}

func assertPanic(t *testing.T, index int, cb func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("failed testCase[%d], panic expected", index)
		}
	}()
	cb()
}

func testWantError(t *testing.T, index int, expectedErr, err error) {
	t.Errorf("failed testCase[%d], error expected '%v' got '%v'", index, expectedErr, err)
}

func testAssertError(t *testing.T, index int, expectedRes, result interface{}) {
	t.Errorf("failed testCase[%d], expected %v got %v", index, expectedRes, result)
}

func testMutateError(t *testing.T, index int, expectedRes, result interface{}) {
	t.Errorf(
		"failed testCase[%d], data has been mutated, before %v after %v",
		index, expectedRes, result,
	)
}
