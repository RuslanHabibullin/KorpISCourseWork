package domain

import "errors"

var (
	// ErrNotFound — запись не найдена
	ErrNotFound = errors.New("not found")

	// ErrNoStock — недостаточно запчастей на складе
	ErrNoStock = errors.New("insufficient stock")

	// ErrInvalidTransition — недопустимый переход статуса FSM
	ErrInvalidTransition = errors.New("invalid status transition")

	// ErrAlreadyExists — запись уже существует
	ErrAlreadyExists = errors.New("already exists")

	// ErrOrderClosed — заказ закрыт, изменения невозможны
	ErrOrderClosed = errors.New("order is closed")
)
