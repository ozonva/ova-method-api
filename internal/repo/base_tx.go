package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txWithContext struct {
	*sqlx.Tx
}

func newTxWithContext(tx *sqlx.Tx) Connection {
	return &txWithContext{tx}
}

func (tx *txWithContext) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return sqlx.NamedQueryContext(ctx, tx, query, arg)
}
