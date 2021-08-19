package flusher

import (
	"log"

	"ova-method-api/internal"
	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
)

type Flusher interface {
	Flush(items []model.Method) []model.Method
}

type flusher struct {
	chunkSize  int
	methodRepo repo.MethodRepo
}

func New(chunkSize int, methodRepo repo.MethodRepo) Flusher {
	return &flusher{
		chunkSize:  chunkSize,
		methodRepo: methodRepo,
	}
}

func (f *flusher) Flush(items []model.Method) []model.Method {
	chunkedItems, err := internal.ListOfMethodToChunkSlice(items, f.chunkSize)
	if err != nil {
		log.Println(err)
		return items
	}

	result := make([]model.Method, 0, len(items))
	for _, chunk := range chunkedItems {
		if err = f.methodRepo.Add(chunk); err != nil {
			log.Println(err)
			result = append(result, chunk...)
		}
	}

	return result
}
