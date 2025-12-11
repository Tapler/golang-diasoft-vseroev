package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
)

// App представляет бизнес-логику приложения календаря.
type App struct {
	logger  Logger
	storage Storage
}

// EventParams содержит параметры события.
type EventParams struct {
	ID           string
	Title        string
	StartTime    time.Time
	Duration     time.Duration
	Description  string
	UserID       int64
	NotifyBefore time.Duration
}

// Logger определяет интерфейс для логирования.
type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

// Storage определяет интерфейс для работы с хранилищем событий.
type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventByID(ctx context.Context, id string) (*storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error)
}

// New создает новый экземпляр приложения.
func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

// CreateEvent создает новое событие в календаре.
func (a *App) CreateEvent(params EventParams) error {
	a.logger.Debug(fmt.Sprintf("Creating event: id=%s, title=%s, user_id=%d",
		params.ID, params.Title, params.UserID))

	event := storage.Event{
		ID:           params.ID,
		Title:        params.Title,
		StartTime:    params.StartTime,
		Duration:     params.Duration,
		Description:  params.Description,
		UserID:       params.UserID,
		NotifyBefore: params.NotifyBefore,
	}

	ctx := context.Background()
	if err := a.storage.CreateEvent(ctx, event); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to create event: %v", err))
		return fmt.Errorf("failed to create event: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Event created successfully: id=%s", params.ID))
	return nil
}

// UpdateEvent обновляет существующее событие.
func (a *App) UpdateEvent(params EventParams) error {
	a.logger.Debug(fmt.Sprintf("Updating event: id=%s, title=%s", params.ID, params.Title))

	event := storage.Event{
		ID:           params.ID,
		Title:        params.Title,
		StartTime:    params.StartTime,
		Duration:     params.Duration,
		Description:  params.Description,
		UserID:       params.UserID,
		NotifyBefore: params.NotifyBefore,
	}

	ctx := context.Background()
	if err := a.storage.UpdateEvent(ctx, params.ID, event); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to update event: %v", err))
		return fmt.Errorf("failed to update event: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Event updated successfully: id=%s", params.ID))
	return nil
}

// DeleteEvent удаляет событие по ID.
func (a *App) DeleteEvent(id string) error {
	a.logger.Debug(fmt.Sprintf("Deleting event: id=%s", id))

	ctx := context.Background()
	if err := a.storage.DeleteEvent(ctx, id); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to delete event: %v", err))
		return fmt.Errorf("failed to delete event: %w", err)
	}

	a.logger.Info(fmt.Sprintf("Event deleted successfully: id=%s", id))
	return nil
}

// GetEventByID возвращает событие по его ID.
func (a *App) GetEventByID(id string) (*storage.Event, error) {
	a.logger.Debug(fmt.Sprintf("Getting event: id=%s", id))

	ctx := context.Background()
	event, err := a.storage.GetEventByID(ctx, id)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to get event: %v", err))
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// ListEventsForDay возвращает список событий на указанную дату.
func (a *App) ListEventsForDay(date time.Time) ([]storage.Event, error) {
	a.logger.Debug(fmt.Sprintf("Listing events for day: %s", date.Format("2006-01-02")))

	ctx := context.Background()
	events, err := a.storage.ListEventsForDay(ctx, date)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to list events for day: %v", err))
		return nil, fmt.Errorf("failed to list events for day: %w", err)
	}

	a.logger.Debug(fmt.Sprintf("Found %d events for day", len(events)))
	return events, nil
}

// ListEventsForWeek возвращает список событий на неделю начиная с указанной даты.
func (a *App) ListEventsForWeek(startDate time.Time) ([]storage.Event, error) {
	a.logger.Debug(fmt.Sprintf("Listing events for week starting: %s", startDate.Format("2006-01-02")))

	ctx := context.Background()
	events, err := a.storage.ListEventsForWeek(ctx, startDate)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to list events for week: %v", err))
		return nil, fmt.Errorf("failed to list events for week: %w", err)
	}

	a.logger.Debug(fmt.Sprintf("Found %d events for week", len(events)))
	return events, nil
}

// ListEventsForMonth возвращает список событий на месяц начиная с указанной даты.
func (a *App) ListEventsForMonth(startDate time.Time) ([]storage.Event, error) {
	a.logger.Debug(fmt.Sprintf("Listing events for month starting: %s", startDate.Format("2006-01-02")))

	ctx := context.Background()
	events, err := a.storage.ListEventsForMonth(ctx, startDate)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to list events for month: %v", err))
		return nil, fmt.Errorf("failed to list events for month: %w", err)
	}

	a.logger.Debug(fmt.Sprintf("Found %d events for month", len(events)))
	return events, nil
}
