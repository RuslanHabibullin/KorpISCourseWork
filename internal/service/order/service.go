package order

import (
	"context"
	"fmt"
	"time"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"

	"github.com/google/uuid"
)

// TxManager — интерфейс для управления транзакциями
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Service — бизнес-логика заказ-нарядов
type Service struct {
	orders   domain.OrderRepository
	stock    domain.StockRepository
	parts    domain.PartRepository
	services domain.ServiceCatalogRepository
	txm      TxManager
}

func NewService(
	orders domain.OrderRepository,
	stock domain.StockRepository,
	parts domain.PartRepository,
	services domain.ServiceCatalogRepository,
	txm TxManager,
) *Service {
	return &Service{
		orders:   orders,
		stock:    stock,
		parts:    parts,
		services: services,
		txm:      txm,
	}
}

// --- Orders ---

type CreateOrderInput struct {
	VehicleID uuid.UUID
	ClientID  uuid.UUID
	Complaint string
}

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.WorkOrder, error) {
	o := &domain.WorkOrder{
		OrderID:     uuid.New(),
		VehicleID:   in.VehicleID,
		ClientID:    in.ClientID,
		Status:      domain.OrderStatusDraft,
		Complaint:   in.Complaint,
		TotalAmount: 0,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.orders.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("CreateOrder: %w", err)
	}
	return o, nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*domain.WorkOrder, error) {
	return s.orders.GetByID(ctx, id)
}

func (s *Service) ListOrders(ctx context.Context, limit, offset int) ([]*domain.WorkOrder, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.orders.List(ctx, limit, offset)
}

// TransitionStatus — FSM переход статуса заказа
func (s *Service) TransitionStatus(ctx context.Context, orderID uuid.UUID, next domain.OrderStatus) (*domain.WorkOrder, error) {
	o, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if !o.Status.CanTransition(next) {
		return nil, fmt.Errorf("%w: %s → %s", domain.ErrInvalidTransition, o.Status, next)
	}
	if err := s.orders.UpdateStatus(ctx, orderID, next); err != nil {
		return nil, fmt.Errorf("TransitionStatus: %w", err)
	}
	o.Status = next
	return o, nil
}

// --- Services (работы) ---

type AddServiceInput struct {
	OrderID   uuid.UUID
	ServiceID uuid.UUID
	Quantity  int
}

func (s *Service) AddService(ctx context.Context, in AddServiceInput) (*domain.WorkOrderService, error) {
	o, err := s.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if o.Status == domain.OrderStatusClosed {
		return nil, domain.ErrOrderClosed
	}

	svc, err := s.services.GetByID(ctx, in.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("AddService: service not found: %w", err)
	}

	ws := &domain.WorkOrderService{
		ID:        uuid.New(),
		OrderID:   in.OrderID,
		ServiceID: in.ServiceID,
		Price:     svc.BasePrice,
		Quantity:  in.Quantity,
	}

	if err := s.orders.AddService(ctx, ws); err != nil {
		return nil, fmt.Errorf("AddService: %w", err)
	}

	if err := s.recalcTotal(ctx, in.OrderID); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *Service) ListServices(ctx context.Context, orderID uuid.UUID) ([]*domain.WorkOrderService, error) {
	return s.orders.ListServices(ctx, orderID)
}

func (s *Service) RemoveService(ctx context.Context, orderID, serviceLineID uuid.UUID) error {
	if err := s.orders.RemoveService(ctx, serviceLineID); err != nil {
		return err
	}
	return s.recalcTotal(ctx, orderID)
}

// --- Parts (запчасти) ---

type AddPartInput struct {
	OrderID  uuid.UUID
	PartID   uuid.UUID
	Quantity int
}

// AddPart добавляет запчасть в заказ и резервирует её на складе
func (s *Service) AddPart(ctx context.Context, in AddPartInput) (*domain.WorkOrderPart, error) {
	o, err := s.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if o.Status == domain.OrderStatusClosed {
		return nil, domain.ErrOrderClosed
	}

	part, err := s.parts.GetByID(ctx, in.PartID)
	if err != nil {
		return nil, fmt.Errorf("AddPart: part not found: %w", err)
	}

	var wp *domain.WorkOrderPart
	err = s.txm.WithTx(ctx, func(ctx context.Context) error {
		// Резервируем на складе
		if err := s.stock.Reserve(ctx, in.PartID, in.Quantity); err != nil {
			return err
		}
		wp = &domain.WorkOrderPart{
			ID:       uuid.New(),
			OrderID:  in.OrderID,
			PartID:   in.PartID,
			Quantity: in.Quantity,
			Price:    part.Price,
		}
		return s.orders.AddPart(ctx, wp)
	})
	if err != nil {
		return nil, fmt.Errorf("AddPart: %w", err)
	}

	if err := s.recalcTotal(ctx, in.OrderID); err != nil {
		return nil, err
	}
	return wp, nil
}

func (s *Service) ListParts(ctx context.Context, orderID uuid.UUID) ([]*domain.WorkOrderPart, error) {
	return s.orders.ListParts(ctx, orderID)
}

// RemovePart удаляет запчасть и возвращает её на склад
func (s *Service) RemovePart(ctx context.Context, orderID uuid.UUID, partLine *domain.WorkOrderPart) error {
	err := s.txm.WithTx(ctx, func(ctx context.Context) error {
		if err := s.stock.Release(ctx, partLine.PartID, partLine.Quantity); err != nil {
			return err
		}
		return s.orders.RemovePart(ctx, partLine.ID)
	})
	if err != nil {
		return fmt.Errorf("RemovePart: %w", err)
	}
	return s.recalcTotal(ctx, orderID)
}

// --- Payments ---

type AddPaymentInput struct {
	OrderID uuid.UUID
	Amount  float64
}

// AddPayment принимает оплату и автоматически закрывает заказ при полной оплате
func (s *Service) AddPayment(ctx context.Context, in AddPaymentInput) (*domain.Payment, error) {
	o, err := s.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if o.Status == domain.OrderStatusClosed {
		return nil, domain.ErrOrderClosed
	}

	p := &domain.Payment{
		PaymentID: uuid.New(),
		OrderID:   in.OrderID,
		Amount:    in.Amount,
		PaidAt:    time.Now().UTC(),
	}

	err = s.txm.WithTx(ctx, func(ctx context.Context) error {
		if err := s.orders.AddPayment(ctx, p); err != nil {
			return err
		}

		// Проверяем сумму платежей
		total, err := s.orders.SumPayments(ctx, in.OrderID)
		if err != nil {
			return err
		}

		// Авто-закрытие при полной оплате
		if total >= o.TotalAmount && o.TotalAmount > 0 {
			return s.orders.UpdateStatus(ctx, in.OrderID, domain.OrderStatusClosed)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("AddPayment: %w", err)
	}
	return p, nil
}

func (s *Service) ListPayments(ctx context.Context, orderID uuid.UUID) ([]*domain.Payment, error) {
	return s.orders.ListPayments(ctx, orderID)
}

// --- Catalog ---

func (s *Service) CreateServiceCatalog(ctx context.Context, name string, price float64) (*domain.Service, error) {
	svc := &domain.Service{
		ServiceID: uuid.New(),
		Name:      name,
		BasePrice: price,
	}
	if err := s.services.Create(ctx, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *Service) ListServiceCatalog(ctx context.Context) ([]*domain.Service, error) {
	return s.services.List(ctx)
}

// --- Internal helpers ---

// recalcTotal пересчитывает total_amount заказа
func (s *Service) recalcTotal(ctx context.Context, orderID uuid.UUID) error {
	svcs, err := s.orders.ListServices(ctx, orderID)
	if err != nil {
		return err
	}
	parts, err := s.orders.ListParts(ctx, orderID)
	if err != nil {
		return err
	}

	var total float64
	for _, ws := range svcs {
		total += ws.Price * float64(ws.Quantity)
	}
	for _, wp := range parts {
		total += wp.Price * float64(wp.Quantity)
	}

	return s.orders.UpdateTotalAmount(ctx, orderID, total)
}

// Ensure TxManager satisfies interface at compile time
var _ TxManager = (*repository.TxManager)(nil)
