package transaction

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Manager interface {
		InTransaction(ctx context.Context, callbacks ...func(ctx context.Context) error) error
	}

	SQLManager struct {
		db  *pgxpool.Pool
		log *slog.Logger
	}

	StubManager struct{}

	contextKey string

	QueryExecutor interface {
		Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
		SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	}
)

const (
	txKey contextKey = "tx"
)

func New(db *pgxpool.Pool, log *slog.Logger) *SQLManager {
	return &SQLManager{db: db, log: log}
}

func (m *SQLManager) With(ctx context.Context) QueryExecutor {
	if val, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return val
	}

	return m.db
}

func (m *SQLManager) InTransaction(ctx context.Context, callbacks ...func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("db.BeginTx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback(ctx)

			panic(p)
		}
	}()

	ctx = context.WithValue(ctx, txKey, tx)

	for _, cb := range callbacks {
		if cbErr := cb(ctx); cbErr != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				m.log.ErrorContext(ctx, "error rolling back a transaction", "err", rollbackErr)
			}

			return cbErr
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func NewStubManager() *StubManager {
	return &StubManager{}
}

func (m *StubManager) InTransaction(ctx context.Context, callbacks ...func(ctx context.Context) error) error {
	for _, cb := range callbacks {
		if err := cb(ctx); err != nil {
			return fmt.Errorf("error in transaction: %w", err)
		}
	}

	return nil
}
