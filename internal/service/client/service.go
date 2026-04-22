package client

import (
	"context"
	"fmt"
	"time"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"

	"github.com/google/uuid"
)

// Service — бизнес-логика клиентов
type Service struct {
	clients  domain.ClientRepository
	vehicles domain.VehicleRepository
}

func NewService(clients domain.ClientRepository, vehicles domain.VehicleRepository) *Service {
	return &Service{clients: clients, vehicles: vehicles}
}

// --- Clients ---

type CreateClientInput struct {
	FullName string
	Phone    string
}

func (s *Service) CreateClient(ctx context.Context, in CreateClientInput) (*domain.Client, error) {
	c := &domain.Client{
		ClientID:  uuid.New(),
		FullName:  in.FullName,
		Phone:     in.Phone,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.clients.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("CreateClient: %w", err)
	}
	return c, nil
}

func (s *Service) GetClient(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	return s.clients.GetByID(ctx, id)
}

func (s *Service) ListClients(ctx context.Context, limit, offset int) ([]*domain.Client, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.clients.List(ctx, limit, offset)
}

type UpdateClientInput struct {
	ClientID uuid.UUID
	FullName string
	Phone    string
}

func (s *Service) UpdateClient(ctx context.Context, in UpdateClientInput) (*domain.Client, error) {
	c, err := s.clients.GetByID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	c.FullName = in.FullName
	c.Phone = in.Phone
	if err := s.clients.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("UpdateClient: %w", err)
	}
	return c, nil
}

func (s *Service) DeleteClient(ctx context.Context, id uuid.UUID) error {
	return s.clients.Delete(ctx, id)
}

// --- Vehicles ---

type CreateVehicleInput struct {
	ClientID uuid.UUID
	Brand    string
	Model    string
	Plate    string
}

func (s *Service) CreateVehicle(ctx context.Context, in CreateVehicleInput) (*domain.Vehicle, error) {
	// Убедиться, что клиент существует
	if _, err := s.clients.GetByID(ctx, in.ClientID); err != nil {
		return nil, fmt.Errorf("CreateVehicle: client not found: %w", err)
	}
	v := &domain.Vehicle{
		VehicleID: uuid.New(),
		ClientID:  in.ClientID,
		Brand:     in.Brand,
		Model:     in.Model,
		Plate:     in.Plate,
	}
	if err := s.vehicles.Create(ctx, v); err != nil {
		return nil, fmt.Errorf("CreateVehicle: %w", err)
	}
	return v, nil
}

func (s *Service) GetVehicle(ctx context.Context, id uuid.UUID) (*domain.Vehicle, error) {
	return s.vehicles.GetByID(ctx, id)
}

func (s *Service) ListVehiclesByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.Vehicle, error) {
	return s.vehicles.ListByClient(ctx, clientID)
}

func (s *Service) UpdateVehicle(ctx context.Context, v *domain.Vehicle) (*domain.Vehicle, error) {
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, fmt.Errorf("UpdateVehicle: %w", err)
	}
	return v, nil
}

func (s *Service) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	return s.vehicles.Delete(ctx, id)
}
