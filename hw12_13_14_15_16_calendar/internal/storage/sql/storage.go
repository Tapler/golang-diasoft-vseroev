package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
	// Импорт драйвера PostgreSQL.
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return &Storage{db: db}, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if event.ID == "" || event.Title == "" {
		return storage.ErrInvalidEvent
	}

	// Проверка на занятость времени
	busy, err := s.isTimeBusy(ctx, event.UserID, event.StartTime, event.Duration, "")
	if err != nil {
		return fmt.Errorf("failed to check time availability: %w", err)
	}
	if busy {
		return storage.ErrDateBusy
	}

	query := `
		INSERT INTO events (id, title, start_time, duration, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	dbEvent := eventToDBEvent(&event)
	_, err = s.db.ExecContext(ctx, query,
		dbEvent.ID,
		dbEvent.Title,
		dbEvent.StartTime,
		dbEvent.Duration,
		dbEvent.Description,
		dbEvent.UserID,
		dbEvent.NotifyBefore,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	if event.Title == "" {
		return storage.ErrInvalidEvent
	}

	// Проверка существования события
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id)
	if err != nil {
		return fmt.Errorf("failed to check event existence: %w", err)
	}
	if !exists {
		return storage.ErrEventNotFound
	}

	// Проверка на занятость времени (исключая текущее событие)
	busy, err := s.isTimeBusy(ctx, event.UserID, event.StartTime, event.Duration, id)
	if err != nil {
		return fmt.Errorf("failed to check time availability: %w", err)
	}
	if busy {
		return storage.ErrDateBusy
	}

	query := `
		UPDATE events
		SET title = $1, start_time = $2, duration = $3, description = $4, user_id = $5, notify_before = $6
		WHERE id = $7
	`

	dbEvent := eventToDBEvent(&event)
	_, err = s.db.ExecContext(ctx, query,
		dbEvent.Title,
		dbEvent.StartTime,
		dbEvent.Duration,
		dbEvent.Description,
		dbEvent.UserID,
		dbEvent.NotifyBefore,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) GetEventByID(ctx context.Context, id string) (*storage.Event, error) {
	var dbEvent dbEvent
	err := s.db.GetContext(ctx, &dbEvent, "SELECT * FROM events WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return dbEvent.toEvent(), nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.listEventsByTimeRange(ctx, startOfDay, endOfDay)
}

func (s *Storage) ListEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.listEventsByTimeRange(ctx, startOfWeek, endOfWeek)
}

func (s *Storage) ListEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startOfMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.listEventsByTimeRange(ctx, startOfMonth, endOfMonth)
}

// Вспомогательные методы

func (s *Storage) isTimeBusy(
	ctx context.Context,
	userID int64,
	startTime time.Time,
	duration time.Duration,
	excludeID string,
) (bool, error) {
	endTime := startTime.Add(duration)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM events
			WHERE user_id = $1
			AND id != $2
			AND start_time < $3
			AND (start_time + duration * INTERVAL '1 second') > $4
		)
	`

	var busy bool
	err := s.db.GetContext(ctx, &busy, query, userID, excludeID, endTime, startTime)
	if err != nil {
		return false, err
	}

	return busy, nil
}

func (s *Storage) listEventsByTimeRange(ctx context.Context, start, end time.Time) ([]storage.Event, error) {
	query := `
		SELECT * FROM events
		WHERE start_time >= $1 AND start_time < $2
		ORDER BY start_time
	`

	var dbEvents []dbEvent
	err := s.db.SelectContext(ctx, &dbEvents, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]storage.Event, 0, len(dbEvents))
	for _, dbEvent := range dbEvents {
		events = append(events, *dbEvent.toEvent())
	}

	return events, nil
}

// dbEvent представляет событие в БД (с типами, совместимыми с SQL).
type dbEvent struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	StartTime    time.Time `db:"start_time"`
	Duration     int64     `db:"duration"`
	Description  string    `db:"description"`
	UserID       int64     `db:"user_id"`
	NotifyBefore int64     `db:"notify_before"`
}

// из БД в доменную модель.
func (d *dbEvent) toEvent() *storage.Event {
	return &storage.Event{
		ID:           d.ID,
		Title:        d.Title,
		StartTime:    d.StartTime,
		Duration:     time.Duration(d.Duration) * time.Second,
		Description:  d.Description,
		UserID:       d.UserID,
		NotifyBefore: time.Duration(d.NotifyBefore) * time.Second,
	}
}

// из доменной модели в БД.
func eventToDBEvent(event *storage.Event) *dbEvent {
	return &dbEvent{
		ID:           event.ID,
		Title:        event.Title,
		StartTime:    event.StartTime,
		Duration:     int64(event.Duration.Seconds()),
		Description:  event.Description,
		UserID:       event.UserID,
		NotifyBefore: int64(event.NotifyBefore.Seconds()),
	}
}
