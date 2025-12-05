package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestStorage_CreateEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	t.Run("successful event creation", func(t *testing.T) {
		event := storage.Event{
			ID:        "event-1",
			Title:     "Meeting",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    1,
		}

		err := s.CreateEvent(ctx, event)
		require.NoError(t, err)

		// Проверяем, что событие создано
		retrieved, err := s.GetEventByID(ctx, "event-1")
		require.NoError(t, err)
		require.Equal(t, event.ID, retrieved.ID)
		require.Equal(t, event.Title, retrieved.Title)
	})

	t.Run("error when ID is empty", func(t *testing.T) {
		event := storage.Event{
			Title:     "Meeting",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    1,
		}

		err := s.CreateEvent(ctx, event)
		require.ErrorIs(t, err, storage.ErrInvalidEvent)
	})

	t.Run("error when Title is empty", func(t *testing.T) {
		event := storage.Event{
			ID:        "event-2",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    1,
		}

		err := s.CreateEvent(ctx, event)
		require.ErrorIs(t, err, storage.ErrInvalidEvent)
	})

	t.Run("error when time slot is busy", func(t *testing.T) {
		startTime := time.Now().Add(24 * time.Hour)

		event1 := storage.Event{
			ID:        "event-3",
			Title:     "Meeting 1",
			StartTime: startTime,
			Duration:  time.Hour,
			UserID:    2,
		}

		err := s.CreateEvent(ctx, event1)
		require.NoError(t, err)

		// Пытаемся создать пересекающееся событие для того же пользователя
		event2 := storage.Event{
			ID:        "event-4",
			Title:     "Meeting 2",
			StartTime: startTime.Add(30 * time.Minute), // Пересечение
			Duration:  time.Hour,
			UserID:    2,
		}

		err = s.CreateEvent(ctx, event2)
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("different users can have events at same time", func(t *testing.T) {
		startTime := time.Now().Add(48 * time.Hour)

		event1 := storage.Event{
			ID:        "event-5",
			Title:     "Meeting User 3",
			StartTime: startTime,
			Duration:  time.Hour,
			UserID:    3,
		}

		err := s.CreateEvent(ctx, event1)
		require.NoError(t, err)

		event2 := storage.Event{
			ID:        "event-6",
			Title:     "Meeting User 4",
			StartTime: startTime,
			Duration:  time.Hour,
			UserID:    4,
		}

		err = s.CreateEvent(ctx, event2)
		require.NoError(t, err)
	})
}

func TestStorage_UpdateEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	t.Run("successful event update", func(t *testing.T) {
		event := storage.Event{
			ID:        "event-10",
			Title:     "Original Title",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    10,
		}

		err := s.CreateEvent(ctx, event)
		require.NoError(t, err)

		updatedEvent := storage.Event{
			Title:     "Updated Title",
			StartTime: event.StartTime,
			Duration:  time.Hour,
			UserID:    10,
		}

		err = s.UpdateEvent(ctx, "event-10", updatedEvent)
		require.NoError(t, err)

		retrieved, err := s.GetEventByID(ctx, "event-10")
		require.NoError(t, err)
		require.Equal(t, "Updated Title", retrieved.Title)
	})

	t.Run("error when event does not exist", func(t *testing.T) {
		event := storage.Event{
			Title:     "Title",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    11,
		}

		err := s.UpdateEvent(ctx, "non-existent", event)
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})
}

func TestStorage_DeleteEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	t.Run("successful event deletion", func(t *testing.T) {
		event := storage.Event{
			ID:        "event-20",
			Title:     "To Delete",
			StartTime: time.Now(),
			Duration:  time.Hour,
			UserID:    20,
		}

		err := s.CreateEvent(ctx, event)
		require.NoError(t, err)

		err = s.DeleteEvent(ctx, "event-20")
		require.NoError(t, err)

		_, err = s.GetEventByID(ctx, "event-20")
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})

	t.Run("error when deleting non-existent event", func(t *testing.T) {
		err := s.DeleteEvent(ctx, "non-existent")
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})
}

func TestStorage_ListEvents(t *testing.T) {
	ctx := context.Background()
	s := New()

	baseTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	// Создаем события в разные дни
	events := []storage.Event{
		{ID: "e1", Title: "Day 1", StartTime: baseTime, Duration: time.Hour, UserID: 15},
		{ID: "e2", Title: "Day 2", StartTime: baseTime.Add(24 * time.Hour), Duration: time.Hour, UserID: 15},
		{ID: "e3", Title: "Day 8", StartTime: baseTime.Add(7 * 24 * time.Hour), Duration: time.Hour, UserID: 15},
		{ID: "e4", Title: "Next Month", StartTime: baseTime.AddDate(0, 1, 0), Duration: time.Hour, UserID: 15},
	}

	for _, event := range events {
		err := s.CreateEvent(ctx, event)
		require.NoError(t, err)
	}

	t.Run("list events for day", func(t *testing.T) {
		dayEvents, err := s.ListEventsForDay(ctx, baseTime)
		require.NoError(t, err)
		require.Len(t, dayEvents, 1)
		require.Equal(t, "Day 1", dayEvents[0].Title)
	})

	t.Run("list events for week", func(t *testing.T) {
		weekEvents, err := s.ListEventsForWeek(ctx, baseTime)
		require.NoError(t, err)
		require.Len(t, weekEvents, 2)
	})

	t.Run("list events for month", func(t *testing.T) {
		monthEvents, err := s.ListEventsForMonth(ctx, baseTime)
		require.NoError(t, err)
		require.Len(t, monthEvents, 3)
	})
}
