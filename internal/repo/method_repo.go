package repo

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"ova-method-api/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/method_repo.go -package=mock

var (
	ErrNoRows = fmt.Errorf("no rows in result set")
)

type MethodRepo interface {
	Add(items []model.Method) error
	Remove(id uint64) error
	List(limit, offset uint64) ([]model.Method, error)
	Describe(id uint64) (*model.Method, error)
}

type methodRepo struct {
	conn *sqlx.DB
}

func NewMethodRepo(conn *sqlx.DB) MethodRepo {
	return &methodRepo{conn: conn}
}

func (rep *methodRepo) Add(items []model.Method) error {
	_, err := rep.conn.NamedExec("INSERT INTO methods (user_id,value) VALUES(:user_id,:value)", items)
	if err != nil {
		return err
	}

	return nil
}

func (rep *methodRepo) Remove(id uint64) error {
	_, err := rep.conn.Exec("DELETE FROM methods WHERE id=$1", id)
	if err != nil {
		return err
	}

	return nil
}

func (rep *methodRepo) List(limit, offset uint64) ([]model.Method, error) {
	var result []model.Method
	err := rep.conn.Select(
		&result,
		"SELECT * FROM methods ORDER BY id LIMIT $1 OFFSET $2",
		limit,
		offset,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rep *methodRepo) Describe(id uint64) (*model.Method, error) {
	var result model.Method
	err := rep.conn.Get(&result, "select * from methods WHERE id=$1", id)

	if err == sql.ErrNoRows {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}
