package storage

import (
	"context"
	"errors"
	"time"
)

// Бизнес-ошибки хранилища.
var (
	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date is busy")
	ErrInvalidEvent  = errors.New("invalid event data")
)

// Storage определяет интерфейс для работы с хранилищем событий.
type Storage interface {
	// CreateEvent добавляет новое событие в хранилище.
	CreateEvent(ctx context.Context, event Event) error

	// UpdateEvent обновляет существующее событие.
	UpdateEvent(ctx context.Context, id string, event Event) error

	// DeleteEvent удаляет событие по ID.
	DeleteEvent(ctx context.Context, id string) error

	// GetEventByID возвращает событие по ID.
	GetEventByID(ctx context.Context, id string) (*Event, error)

	// ListEventsForDay возвращает список событий на указанную дату.
	ListEventsForDay(ctx context.Context, date time.Time) ([]Event, error)

	// ListEventsForWeek возвращает список событий на неделю начиная с указанной даты.
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]Event, error)

	// ListEventsForMonth возвращает список событий на месяц начиная с указанной даты.
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]Event, error)
}
