package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const pgxKey = "pgxKey"

type Transactor interface {
	Transact(ctx context.Context, txFn func(context.Context) error) error
}

type transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *transactor {
	return &transactor{
		pool: pool,
	}
}

func (t *transactor) Transact(ctx context.Context, txFn func(context.Context) error) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginTx: %w", err)
	}
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = txFn(injectTx(ctx, tx))
	if err != nil {
		return fmt.Errorf("txFn: %w", err)
	}
	return err
}

type commonPgx interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, pgxKey, tx)
}

func (t *transactor) extractTx(ctx context.Context) commonPgx {
	if tx, ok := ctx.Value(pgxKey).(pgx.Tx); ok {
		return tx
	}
	return t.pool
}
