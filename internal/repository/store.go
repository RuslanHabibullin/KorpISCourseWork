package repository

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type ctxKey string

const txKey ctxKey = "tx"

// Store хранит соединение с БД
type Store struct {
	DB *sqlx.DB
}

// NewStore создаёт новое соединение с PostgreSQL
func NewStore(dsn string) (*Store, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	return &Store{DB: db}, nil
}

// TxManager управляет транзакциями
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager создаёт новый TxManager
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx выполняет fn внутри транзакции, пробрасывая её через context
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	ctx = context.WithValue(ctx, txKey, tx)

	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// ExtractTx извлекает транзакцию из context (если есть)
func ExtractTx(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txKey).(*sqlx.Tx)
	return tx
}

// Queryer возвращает либо tx, либо db — для использования в репозиториях
func Queryer(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return db
}
