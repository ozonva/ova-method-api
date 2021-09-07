package repo

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Connection interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, args interface{}) (sql.Result, error)
	NamedQuery(query string, args interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

type Transactionable interface {
	Connection

	Beginx() (*sqlx.Tx, error)
}

type baseRepo struct {
	conn Connection
}

func newBaseRepo(conn Connection) baseRepo {
	return baseRepo{conn}
}

func (rep *baseRepo) Transaction(fn func(tx *sqlx.Tx) error) error {
	txConn, ok := rep.conn.(Transactionable)
	if !ok {
		return fmt.Errorf("transactions are not supported")
	}

	tx, err := txConn.Beginx()
	if err != nil {
		return err
	}

	if err = fn(tx); err != nil {
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
