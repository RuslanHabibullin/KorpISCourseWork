package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OrderStatus — статус заказ-наряда
type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusApproved   OrderStatus = "approved"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusDone       OrderStatus = "done"
	OrderStatusClosed     OrderStatus = "closed"
)

// validTransitions — разрешённые переходы FSM
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusDraft:      {OrderStatusApproved},
	OrderStatusApproved:   {OrderStatusInProgress},
	OrderStatusInProgress: {OrderStatusDone},
	OrderStatusDone:       {OrderStatusClosed},
	OrderStatusClosed:     {},
}

// CanTransition проверяет, допустим ли переход между статусами
func (s OrderStatus) CanTransition(next OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

// WorkOrder — заказ-наряд
type WorkOrder struct {
	OrderID     uuid.UUID   `db:"order_id"     json:"order_id"`
	VehicleID   uuid.UUID   `db:"vehicle_id"   json:"vehicle_id"`
	ClientID    uuid.UUID   `db:"client_id"    json:"client_id"`
	Status      OrderStatus `db:"status"       json:"status"`
	Complaint   string      `db:"complaint"    json:"complaint"`
	TotalAmount float64     `db:"total_amount" json:"total_amount"`
	CreatedAt   time.Time   `db:"created_at"   json:"created_at"`
}

// Service — услуга автосервиса
type Service struct {
	ServiceID uuid.UUID `db:"service_id" json:"service_id"`
	Name      string    `db:"name"       json:"name"`
	BasePrice float64   `db:"base_price" json:"base_price"`
}

// WorkOrderService — услуга в заказ-наряде
type WorkOrderService struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	OrderID   uuid.UUID `db:"order_id"   json:"order_id"`
	ServiceID uuid.UUID `db:"service_id" json:"service_id"`
	Price     float64   `db:"price"      json:"price"`
	Quantity  int       `db:"quantity"   json:"quantity"`
}

// Part — запчасть
type Part struct {
	PartID uuid.UUID `db:"part_id" json:"part_id"`
	Name   string    `db:"name"    json:"name"`
	Price  float64   `db:"price"   json:"price"`
}

// WorkOrderPart — запчасть в заказ-наряде
type WorkOrderPart struct {
	ID       uuid.UUID `db:"id"       json:"id"`
	OrderID  uuid.UUID `db:"order_id" json:"order_id"`
	PartID   uuid.UUID `db:"part_id"  json:"part_id"`
	Quantity int       `db:"quantity" json:"quantity"`
	Price    float64   `db:"price"    json:"price"`
}

// Payment — платёж по заказ-наряду
type Payment struct {
	PaymentID uuid.UUID `db:"payment_id" json:"payment_id"`
	OrderID   uuid.UUID `db:"order_id"   json:"order_id"`
	Amount    float64   `db:"amount"     json:"amount"`
	PaidAt    time.Time `db:"paid_at"    json:"paid_at"`
}

// OrderRepository — интерфейс репозитория заказ-нарядов
type OrderRepository interface {
	Create(ctx context.Context, o *WorkOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*WorkOrder, error)
	List(ctx context.Context, limit, offset int) ([]*WorkOrder, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
	UpdateTotalAmount(ctx context.Context, id uuid.UUID, total float64) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddService(ctx context.Context, ws *WorkOrderService) error
	ListServices(ctx context.Context, orderID uuid.UUID) ([]*WorkOrderService, error)
	RemoveService(ctx context.Context, id uuid.UUID) error

	AddPart(ctx context.Context, wp *WorkOrderPart) error
	ListParts(ctx context.Context, orderID uuid.UUID) ([]*WorkOrderPart, error)
	RemovePart(ctx context.Context, id uuid.UUID) error

	AddPayment(ctx context.Context, p *Payment) error
	SumPayments(ctx context.Context, orderID uuid.UUID) (float64, error)
	ListPayments(ctx context.Context, orderID uuid.UUID) ([]*Payment, error)
}

// ServiceCatalogRepository — интерфейс репозитория услуг
type ServiceCatalogRepository interface {
	Create(ctx context.Context, s *Service) error
	GetByID(ctx context.Context, id uuid.UUID) (*Service, error)
	List(ctx context.Context) ([]*Service, error)
}

// PartRepository — интерфейс репозитория запчастей
type PartRepository interface {
	Create(ctx context.Context, p *Part) error
	GetByID(ctx context.Context, id uuid.UUID) (*Part, error)
	List(ctx context.Context) ([]*Part, error)
}
