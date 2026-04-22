package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// ErrNoRows — sentinel для "нет строк"
var ErrNoRows = errors.New("no rows in result set")

// isPgError проверяет код ошибки PostgreSQL
func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

type stockRepo struct {
	db *sqlx.DB
}

func NewStockRepository(db *sqlx.DB) domain.StockRepository {
	return &stockRepo{db: db}
}

func (r *stockRepo) GetByPartID(ctx context.Context, partID uuid.UUID) (*domain.Stock, error) {
	var s domain.Stock
	q := `SELECT part_id, qty FROM stock WHERE part_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &s, q, partID)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("stockRepo.GetByPartID: %w", err)
	}
	return &s, nil
}

// Reserve уменьшает остаток; возвращает ErrNoStock если недостаточно
func (r *stockRepo) Reserve(ctx context.Context, partID uuid.UUID, qty int) error {
	q := `UPDATE stock SET qty = qty - $1 WHERE part_id = $2 AND qty >= $1`
	res, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, qty, partID)
	if err != nil {
		return fmt.Errorf("stockRepo.Reserve: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNoStock
	}
	return nil
}

func (r *stockRepo) Release(ctx context.Context, partID uuid.UUID, qty int) error {
	q := `UPDATE stock SET qty = qty + $1 WHERE part_id = $2`
	_, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, qty, partID)
	if err != nil {
		return fmt.Errorf("stockRepo.Release: %w", err)
	}
	return nil
}

func (r *stockRepo) Replenish(ctx context.Context, partID uuid.UUID, qty int) error {
	q := `INSERT INTO stock (part_id, qty) VALUES ($1, $2)
	      ON CONFLICT (part_id) DO UPDATE SET qty = stock.qty + $2`
	_, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, partID, qty)
	if err != nil {
		return fmt.Errorf("stockRepo.Replenish: %w", err)
	}
	return nil
}

func (r *stockRepo) List(ctx context.Context) ([]*domain.Stock, error) {
	var list []*domain.Stock
	q := `SELECT part_id, qty FROM stock ORDER BY part_id`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q)
	if err != nil {
		return nil, fmt.Errorf("stockRepo.List: %w", err)
	}
	return list, nil
}
