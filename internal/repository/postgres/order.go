package postgres

import (
	"context"
	"fmt"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type orderRepo struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, o *domain.WorkOrder) error {
	q := `INSERT INTO work_order (order_id, vehicle_id, client_id, status, complaint, total_amount, created_at)
	      VALUES (:order_id, :vehicle_id, :client_id, :status, :complaint, :total_amount, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, o)
	if err != nil {
		return fmt.Errorf("orderRepo.Create: %w", err)
	}
	return nil
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkOrder, error) {
	var o domain.WorkOrder
	q := `SELECT order_id, vehicle_id, client_id, status, complaint, total_amount, created_at
	      FROM work_order WHERE order_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &o, q, id)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("orderRepo.GetByID: %w", err)
	}
	return &o, nil
}

func (r *orderRepo) List(ctx context.Context, limit, offset int) ([]*domain.WorkOrder, error) {
	var orders []*domain.WorkOrder
	q := `SELECT order_id, vehicle_id, client_id, status, complaint, total_amount, created_at
	      FROM work_order ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &orders, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.List: %w", err)
	}
	return orders, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	q := `UPDATE work_order SET status = $1 WHERE order_id = $2`
	res, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("orderRepo.UpdateStatus: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *orderRepo) UpdateTotalAmount(ctx context.Context, id uuid.UUID, total float64) error {
	q := `UPDATE work_order SET total_amount = $1 WHERE order_id = $2`
	_, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, total, id)
	if err != nil {
		return fmt.Errorf("orderRepo.UpdateTotalAmount: %w", err)
	}
	return nil
}

func (r *orderRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM work_order WHERE order_id = $1`
	res, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("orderRepo.Delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// --- Services ---

func (r *orderRepo) AddService(ctx context.Context, ws *domain.WorkOrderService) error {
	q := `INSERT INTO work_order_service (id, order_id, service_id, price, quantity)
	      VALUES (:id, :order_id, :service_id, :price, :quantity)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, ws)
	if err != nil {
		return fmt.Errorf("orderRepo.AddService: %w", err)
	}
	return nil
}

func (r *orderRepo) ListServices(ctx context.Context, orderID uuid.UUID) ([]*domain.WorkOrderService, error) {
	var list []*domain.WorkOrderService
	q := `SELECT id, order_id, service_id, price, quantity FROM work_order_service WHERE order_id = $1`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.ListServices: %w", err)
	}
	return list, nil
}

func (r *orderRepo) RemoveService(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM work_order_service WHERE id = $1`
	_, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, id)
	return err
}

// --- Parts ---

func (r *orderRepo) AddPart(ctx context.Context, wp *domain.WorkOrderPart) error {
	q := `INSERT INTO work_order_part (id, order_id, part_id, quantity, price)
	      VALUES (:id, :order_id, :part_id, :quantity, :price)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, wp)
	if err != nil {
		return fmt.Errorf("orderRepo.AddPart: %w", err)
	}
	return nil
}

func (r *orderRepo) ListParts(ctx context.Context, orderID uuid.UUID) ([]*domain.WorkOrderPart, error) {
	var list []*domain.WorkOrderPart
	q := `SELECT id, order_id, part_id, quantity, price FROM work_order_part WHERE order_id = $1`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.ListParts: %w", err)
	}
	return list, nil
}

func (r *orderRepo) RemovePart(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM work_order_part WHERE id = $1`
	_, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, id)
	return err
}

// --- Payments ---

func (r *orderRepo) AddPayment(ctx context.Context, p *domain.Payment) error {
	q := `INSERT INTO payment (payment_id, order_id, amount, paid_at)
	      VALUES (:payment_id, :order_id, :amount, :paid_at)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, p)
	if err != nil {
		return fmt.Errorf("orderRepo.AddPayment: %w", err)
	}
	return nil
}

func (r *orderRepo) SumPayments(ctx context.Context, orderID uuid.UUID) (float64, error) {
	var total float64
	q := `SELECT COALESCE(SUM(amount), 0) FROM payment WHERE order_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &total, q, orderID)
	if err != nil {
		return 0, fmt.Errorf("orderRepo.SumPayments: %w", err)
	}
	return total, nil
}

func (r *orderRepo) ListPayments(ctx context.Context, orderID uuid.UUID) ([]*domain.Payment, error) {
	var list []*domain.Payment
	q := `SELECT payment_id, order_id, amount, paid_at FROM payment WHERE order_id = $1`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.ListPayments: %w", err)
	}
	return list, nil
}

// --- ServiceCatalog ---

type serviceCatalogRepo struct {
	db *sqlx.DB
}

func NewServiceCatalogRepository(db *sqlx.DB) domain.ServiceCatalogRepository {
	return &serviceCatalogRepo{db: db}
}

func (r *serviceCatalogRepo) Create(ctx context.Context, s *domain.Service) error {
	q := `INSERT INTO service (service_id, name, base_price) VALUES (:service_id, :name, :base_price)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, s)
	return err
}

func (r *serviceCatalogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	var s domain.Service
	q := `SELECT service_id, name, base_price FROM service WHERE service_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &s, q, id)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *serviceCatalogRepo) List(ctx context.Context) ([]*domain.Service, error) {
	var list []*domain.Service
	q := `SELECT service_id, name, base_price FROM service ORDER BY name`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q)
	return list, err
}

// --- PartRepository ---

type partRepo struct {
	db *sqlx.DB
}

func NewPartRepository(db *sqlx.DB) domain.PartRepository {
	return &partRepo{db: db}
}

func (r *partRepo) Create(ctx context.Context, p *domain.Part) error {
	q := `INSERT INTO part (part_id, name, price) VALUES (:part_id, :name, :price)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, p)
	return err
}

func (r *partRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	var p domain.Part
	q := `SELECT part_id, name, price FROM part WHERE part_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &p, q, id)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *partRepo) List(ctx context.Context) ([]*domain.Part, error) {
	var list []*domain.Part
	q := `SELECT part_id, name, price FROM part ORDER BY name`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &list, q)
	return list, err
}
