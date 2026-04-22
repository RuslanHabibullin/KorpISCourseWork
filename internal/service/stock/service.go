package stock

import (
	"context"
	"fmt"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"

	"github.com/google/uuid"
)

// Service — бизнес-логика склада
type Service struct {
	stock domain.StockRepository
	parts domain.PartRepository
}

func NewService(stock domain.StockRepository, parts domain.PartRepository) *Service {
	return &Service{stock: stock, parts: parts}
}

func (s *Service) GetStock(ctx context.Context, partID uuid.UUID) (*domain.Stock, error) {
	return s.stock.GetByPartID(ctx, partID)
}

func (s *Service) ListStock(ctx context.Context) ([]*domain.Stock, error) {
	return s.stock.List(ctx)
}

func (s *Service) Replenish(ctx context.Context, partID uuid.UUID, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	return s.stock.Replenish(ctx, partID, qty)
}

// CreatePart создаёт новую запчасть и инициализирует запись склада
func (s *Service) CreatePart(ctx context.Context, name string, price float64) (*domain.Part, error) {
	p := &domain.Part{
		PartID: uuid.New(),
		Name:   name,
		Price:  price,
	}
	if err := s.parts.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("CreatePart: %w", err)
	}
	// Инициализируем склад с qty=0
	if err := s.stock.Replenish(ctx, p.PartID, 0); err != nil {
		return nil, fmt.Errorf("CreatePart: init stock: %w", err)
	}
	return p, nil
}

func (s *Service) GetPart(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	return s.parts.GetByID(ctx, id)
}

func (s *Service) ListParts(ctx context.Context) ([]*domain.Part, error) {
	return s.parts.List(ctx)
}
