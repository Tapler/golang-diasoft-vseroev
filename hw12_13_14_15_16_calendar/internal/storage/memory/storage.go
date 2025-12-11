package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" || event.Title == "" {
		return storage.ErrInvalidEvent
	}

	// Проверка на занятость времени для данного пользователя
	if s.isTimeBusyLocked(event.UserID, event.StartTime, event.Duration, "") {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	if event.Title == "" {
		return storage.ErrInvalidEvent
	}

	// Проверка на занятость времени (исключая текущее событие)
	if s.isTimeBusyLocked(event.UserID, event.StartTime, event.Duration, id) {
		return storage.ErrDateBusy
	}

	event.ID = id
	s.events[id] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) GetEventByID(_ context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, storage.ErrEventNotFound
	}

	return &event, nil
}

func (s *Storage) ListEventsForDay(_ context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.filterEventsByTimeRange(startOfDay, endOfDay), nil
}

func (s *Storage) ListEventsForWeek(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfWeek := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.filterEventsByTimeRange(startOfWeek, endOfWeek), nil
}

func (s *Storage) ListEventsForMonth(_ context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.filterEventsByTimeRange(startOfMonth, endOfMonth), nil
}

// isTimeBusyLocked проверяет, занято ли время для пользователя.
func (s *Storage) isTimeBusyLocked(userID int64, startTime time.Time, duration time.Duration, excludeID string) bool {
	endTime := startTime.Add(duration)

	for id, event := range s.events {
		if id == excludeID {
			continue
		}
		if event.UserID != userID {
			continue
		}

		eventEnd := event.StartTime.Add(event.Duration)

		// Проверка пересечения временных интервалов
		if startTime.Before(eventEnd) && endTime.After(event.StartTime) {
			return true
		}
	}

	return false
}

// filterEventsByTimeRange возвращает события в заданном временном диапазоне.
func (s *Storage) filterEventsByTimeRange(start, end time.Time) []storage.Event {
	var result []storage.Event

	for _, event := range s.events {
		if !event.StartTime.Before(start) && event.StartTime.Before(end) {
			result = append(result, event)
		}
	}

	return result
}
