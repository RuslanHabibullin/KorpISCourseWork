package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type clientRepo struct {
	db *sqlx.DB
}

// NewClientRepository создаёт репозиторий клиентов
func NewClientRepository(db *sqlx.DB) domain.ClientRepository {
	return &clientRepo{db: db}
}

func (r *clientRepo) Create(ctx context.Context, c *domain.Client) error {
	q := `INSERT INTO client (client_id, full_name, phone, created_at)
	      VALUES (:client_id, :full_name, :phone, :created_at)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, c)
	if err != nil {
		return fmt.Errorf("clientRepo.Create: %w", err)
	}
	return nil
}

func (r *clientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	var c domain.Client
	q := `SELECT client_id, full_name, phone, created_at FROM client WHERE client_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &c, q, id)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("clientRepo.GetByID: %w", err)
	}
	return &c, nil
}

func (r *clientRepo) List(ctx context.Context, limit, offset int) ([]*domain.Client, error) {
	var clients []*domain.Client
	q := `SELECT client_id, full_name, phone, created_at FROM client ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &clients, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("clientRepo.List: %w", err)
	}
	return clients, nil
}

func (r *clientRepo) Update(ctx context.Context, c *domain.Client) error {
	q := `UPDATE client SET full_name = :full_name, phone = :phone WHERE client_id = :client_id`
	res, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, c)
	if err != nil {
		return fmt.Errorf("clientRepo.Update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *clientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM client WHERE client_id = $1`
	res, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("clientRepo.Delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// VehicleRepository

type vehicleRepo struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) domain.VehicleRepository {
	return &vehicleRepo{db: db}
}

func (r *vehicleRepo) Create(ctx context.Context, v *domain.Vehicle) error {
	q := `INSERT INTO vehicle (vehicle_id, client_id, brand, model, plate)
	      VALUES (:vehicle_id, :client_id, :brand, :model, :plate)`
	_, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, v)
	if err != nil {
		return fmt.Errorf("vehicleRepo.Create: %w", err)
	}
	return nil
}

func (r *vehicleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	var v domain.Vehicle
	q := `SELECT vehicle_id, client_id, brand, model, plate FROM vehicle WHERE vehicle_id = $1`
	err := sqlx.GetContext(ctx, repository.Queryer(ctx, r.db), &v, q, id)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("vehicleRepo.GetByID: %w", err)
	}
	return &v, nil
}

func (r *vehicleRepo) ListByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.Vehicle, error) {
	var vehicles []*domain.Vehicle
	q := `SELECT vehicle_id, client_id, brand, model, plate FROM vehicle WHERE client_id = $1`
	err := sqlx.SelectContext(ctx, repository.Queryer(ctx, r.db), &vehicles, q, clientID)
	if err != nil {
		return nil, fmt.Errorf("vehicleRepo.ListByClient: %w", err)
	}
	return vehicles, nil
}

func (r *vehicleRepo) Update(ctx context.Context, v *domain.Vehicle) error {
	q := `UPDATE vehicle SET brand = :brand, model = :model, plate = :plate WHERE vehicle_id = :vehicle_id`
	res, err := sqlx.NamedExecContext(ctx, repository.Queryer(ctx, r.db), q, v)
	if err != nil {
		return fmt.Errorf("vehicleRepo.Update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *vehicleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM vehicle WHERE vehicle_id = $1`
	res, err := repository.Queryer(ctx, r.db).(sqlx.ExecerContext).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("vehicleRepo.Delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// isNoRows проверяет ошибку "no rows"
func isNoRows(err error) bool {
	return errors.Is(err, nil)
}
