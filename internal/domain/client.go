package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Client — клиент автосервиса
type Client struct {
	ClientID  uuid.UUID `db:"client_id"  json:"client_id"`
	FullName  string    `db:"full_name"  json:"full_name"`
	Phone     string    `db:"phone"      json:"phone"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Vehicle — автомобиль клиента
type Vehicle struct {
	VehicleID uuid.UUID `db:"vehicle_id" json:"vehicle_id"`
	ClientID  uuid.UUID `db:"client_id"  json:"client_id"`
	Brand     string    `db:"brand"      json:"brand"`
	Model     string    `db:"model"      json:"model"`
	Plate     string    `db:"plate"      json:"plate"`
}

// ClientRepository — интерфейс репозитория клиентов
type ClientRepository interface {
	Create(ctx context.Context, c *Client) error
	GetByID(ctx context.Context, id uuid.UUID) (*Client, error)
	List(ctx context.Context, limit, offset int) ([]*Client, error)
	Update(ctx context.Context, c *Client) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// VehicleRepository — интерфейс репозитория автомобилей
type VehicleRepository interface {
	Create(ctx context.Context, v *Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*Vehicle, error)
	ListByClient(ctx context.Context, clientID uuid.UUID) ([]*Vehicle, error)
	Update(ctx context.Context, v *Vehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
}
