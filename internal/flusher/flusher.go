package flusher

import (
	"context"
	"log"

	"ova-method-api/internal"
	"ova-method-api/internal/model"
	"ova-method-api/internal/repo"
)

type Flusher interface {
	Flush(ctx context.Context, items []model.Method) []model.Method
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

func (f *flusher) Flush(ctx context.Context, items []model.Method) []model.Method {
	chunkedItems, err := internal.ListOfMethodToChunkSlice(items, f.chunkSize)
	if err != nil {
		log.Println(err)
		return items
	}

	var result []model.Method
	for _, chunk := range chunkedItems {
		if _, err = f.methodRepo.Add(ctx, chunk); err != nil {
			log.Println(err)
			result = append(result, chunk...)
		}
	}

	return result
}
