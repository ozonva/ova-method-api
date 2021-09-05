package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Connection interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, args interface{}) (sql.Result, error)
	NamedQueryContext(ctx context.Context, query string, args interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Transactionable interface {
	Connection

	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

type baseRepo struct {
	conn Connection
}

func newBaseRepo(conn Connection) baseRepo {
	return baseRepo{conn}
}

func (rep *baseRepo) Transaction(ctx context.Context, fn func(conn Connection) error) error {
	txConn, ok := rep.conn.(Transactionable)
	if !ok {
		return fmt.Errorf("transactions are not supported")
	}

	tx, err := txConn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err = fn(newTxWithContext(tx)); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return errors.Wrapf(err, "tx error %+v with err:", txErr)
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
