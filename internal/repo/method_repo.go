package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	"ova-method-api/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/method_repo.go -package=mock

var (
	ErrNoRows        = fmt.Errorf("no rows in result set")
	ErrNoRowAffected = fmt.Errorf("no rows affected")
)

type MethodRepo interface {
	Add(ctx context.Context, items []model.Method) ([]model.Method, error)
	Update(ctx context.Context, id uint64, value string) error
	Remove(ctx context.Context, id uint64) error
	List(ctx context.Context, limit, offset uint64) ([]model.Method, error)
	Describe(ctx context.Context, id uint64) (*model.Method, error)
	Transaction(ctx context.Context, fn func(rep MethodRepo) error) error
}

type methodRepo struct {
	baseRepo
}

func NewMethodRepo(conn Connection) MethodRepo {
	return &methodRepo{newBaseRepo(conn)}
}

func (rep *methodRepo) Transaction(ctx context.Context, fn func(rep MethodRepo) error) error {
	return rep.baseRepo.Transaction(ctx, func(conn Connection) error {
		return fn(NewMethodRepo(conn))
	})
}

func (rep *methodRepo) Add(ctx context.Context, items []model.Method) ([]model.Method, error) {
	builder := squirrel.
		Insert("methods").
		Columns("user_id", "value").
		Suffix("RETURNING id, user_id, value, created_at").
		PlaceholderFormat(squirrel.Dollar)

	for _, item := range items {
		builder = builder.Values(item.UserId, item.Value)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := rep.conn.QueryxContext(ctx, query, args...)
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

func (rep *methodRepo) Update(ctx context.Context, id uint64, value string) error {
	query, args, err := squirrel.
		Update("methods").
		Set("value", value).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	res, err := rep.conn.ExecContext(ctx, query, args...)
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

func (rep *methodRepo) Remove(ctx context.Context, id uint64) error {
	query, args, err := squirrel.
		Delete("methods").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	res, err := rep.conn.ExecContext(ctx, query, args...)
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

func (rep *methodRepo) List(ctx context.Context, limit, offset uint64) ([]model.Method, error) {
	query, args, err := squirrel.
		Select("*").
		From("methods").
		OrderBy("id asc").
		Limit(limit).
		Offset(offset).
		ToSql()

	if err != nil {
		return nil, err
	}

	var result []model.Method
	err = rep.conn.SelectContext(ctx, &result, query, args...)

	if err == sql.ErrNoRows {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (rep *methodRepo) Describe(ctx context.Context, id uint64) (*model.Method, error) {
	query, args, err := squirrel.
		Select("*").
		From("methods").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	var result model.Method
	err = rep.conn.GetContext(ctx, &result, query, args...)

	if err == sql.ErrNoRows {
		return nil, ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}
