package repo

import (
	"io"
	"io/ioutil"
	"strings"

	"ova-method-api/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/method_repo.go -package=mock

type MethodRepo interface {
	Add(items []model.Method) error
	List(limit, offset uint64) ([]model.Method, error)
	Describe(id uint64) (*model.Method, error)
}

type methodRepo struct {
}

func NewMethodRepo() MethodRepo {
	return &methodRepo{}
}

func (m *methodRepo) Add(items []model.Method) error {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(item.String())
	}

	reader := strings.NewReader(builder.String())
	_, err := io.Copy(ioutil.Discard, reader)
	return err
}

func (m *methodRepo) List(limit, offset uint64) ([]model.Method, error) {
	panic("implement me")
}

func (m *methodRepo) Describe(id uint64) (*model.Method, error) {
	panic("implement me")
}
