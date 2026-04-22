package etl

import (
	"context"
	"fmt"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"

	"github.com/jmoiron/sqlx"
)

// Record — универсальная запись для ETL-пайплайна
type Record struct {
	OrderID     string  `db:"order_id"`
	ClientName  string  `db:"full_name"`
	Phone       string  `db:"phone"`
	Brand       string  `db:"brand"`
	Model       string  `db:"model"`
	Plate       string  `db:"plate"`
	Status      string  `db:"status"`
	Complaint   string  `db:"complaint"`
	TotalAmount float64 `db:"total_amount"`
	PaidTotal   float64 `db:"paid_total"`
}

// Extractor извлекает данные из БД для ETL-отчётов
type Extractor struct {
	db *sqlx.DB
}

func NewExtractor(db *sqlx.DB) *Extractor {
	return &Extractor{db: db}
}

// ExtractOrders извлекает заказы с фильтрацией по статусу
func (e *Extractor) ExtractOrders(ctx context.Context, status domain.OrderStatus) ([]*Record, error) {
	q := `
		SELECT
			wo.order_id,
			c.full_name,
			c.phone,
			v.brand,
			v.model,
			v.plate,
			wo.status,
			wo.complaint,
			wo.total_amount,
			COALESCE(SUM(p.amount), 0) AS paid_total
		FROM work_order wo
		JOIN client  c ON c.client_id  = wo.client_id
		JOIN vehicle v ON v.vehicle_id = wo.vehicle_id
		LEFT JOIN payment p ON p.order_id = wo.order_id
		WHERE wo.status = $1
		GROUP BY wo.order_id, c.full_name, c.phone, v.brand, v.model, v.plate,
		         wo.status, wo.complaint, wo.total_amount
		ORDER BY wo.created_at DESC
	`
	var records []*Record
	if err := e.db.SelectContext(ctx, &records, q, status); err != nil {
		return nil, fmt.Errorf("extractor.ExtractOrders: %w", err)
	}
	return records, nil
}

// ExtractAllOrders извлекает все заказы (без фильтра статуса)
func (e *Extractor) ExtractAllOrders(ctx context.Context) ([]*Record, error) {
	q := `
		SELECT
			wo.order_id,
			c.full_name,
			c.phone,
			v.brand,
			v.model,
			v.plate,
			wo.status,
			wo.complaint,
			wo.total_amount,
			COALESCE(SUM(p.amount), 0) AS paid_total
		FROM work_order wo
		JOIN client  c ON c.client_id  = wo.client_id
		JOIN vehicle v ON v.vehicle_id = wo.vehicle_id
		LEFT JOIN payment p ON p.order_id = wo.order_id
		GROUP BY wo.order_id, c.full_name, c.phone, v.brand, v.model, v.plate,
		         wo.status, wo.complaint, wo.total_amount
		ORDER BY wo.created_at DESC
	`
	var records []*Record
	if err := e.db.SelectContext(ctx, &records, q); err != nil {
		return nil, fmt.Errorf("extractor.ExtractAllOrders: %w", err)
	}
	return records, nil
}
