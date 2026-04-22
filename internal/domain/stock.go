package domain

import (
	"context"

	"github.com/google/uuid"
)

// Stock — запись склада
type Stock struct {
	PartID uuid.UUID `db:"part_id" json:"part_id"`
	Qty    int       `db:"qty"     json:"qty"`
}

// StockRepository — интерфейс репозитория склада
type StockRepository interface {
	GetByPartID(ctx context.Context, partID uuid.UUID) (*Stock, error)
	Reserve(ctx context.Context, partID uuid.UUID, qty int) error // уменьшить остаток
	Release(ctx context.Context, partID uuid.UUID, qty int) error // вернуть остаток
	Replenish(ctx context.Context, partID uuid.UUID, qty int) error
	List(ctx context.Context) ([]*Stock, error)
}
