package repo

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"ova-method-api/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/method_repo.go -package=mock

var (
	ErrNoRows        = fmt.Errorf("no rows in result set")
	ErrNoRowAffected = fmt.Errorf("no rows affected")
)

type MethodRepo interface {
	Add(items []model.Method) ([]model.Method, error)
	Update(id uint64, value string) error
	Remove(id uint64) error
	List(limit, offset uint64) ([]model.Method, error)
	Describe(id uint64) (*model.Method, error)
	Transaction(fn func(rep MethodRepo) error) error
}

type methodRepo struct {
	baseRepo
}

func NewMethodRepo(conn Connection) MethodRepo {
	return &methodRepo{newBaseRepo(conn)}
}

func (rep *methodRepo) Transaction(fn func(rep MethodRepo) error) error {
	return rep.baseRepo.Transaction(func(tx *sqlx.Tx) error {
		return fn(NewMethodRepo(tx))
	})
}

func (rep *methodRepo) Add(items []model.Method) ([]model.Method, error) {
	rows, err := rep.conn.NamedQuery(
		"INSERT INTO methods (user_id,value) VALUES(:user_id,:value) RETURNING id, user_id, value, created_at",
		items,
	)
	if err != nil {
		return nil, err
	}

	withCloseRows := func(err error) error {
		if closeErr := rows.Close(); closeErr != nil {
			return errors.Wrap(err, "failed close db query rows")
		}
		return err
	}

	result := make([]model.Method, 0, len(items))
	for rows.Next() {
		var method model.Method
		if err = rows.StructScan(&method); err != nil {
			return nil, withCloseRows(err)
		}
		result = append(result, method)
	}

	if err = rows.Err(); err != nil {
		return nil, withCloseRows(err)
	}

	return result, withCloseRows(nil)
}

func (rep *methodRepo) Update(id uint64, value string) error {
	res, err := rep.conn.Exec("update methods set value=$1 where id=$2", value, id)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return ErrNoRowAffected
	}

	return nil
}

func (rep *methodRepo) Remove(id uint64) error {
	res, err := rep.conn.Exec("DELETE FROM methods WHERE id=$1", id)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return ErrNoRowAffected
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
