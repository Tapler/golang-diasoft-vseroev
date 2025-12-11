package internalhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/api"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

const testEventID = "550e8400-e29b-41d4-a716-446655440000"

// MockLogger реализует интерфейс Logger для тестов.
type MockLogger struct{}

func (m *MockLogger) Info(msg string)  {}
func (m *MockLogger) Error(msg string) {}
func (m *MockLogger) Debug(msg string) {}

// MockApp реализует интерфейс CalendarApplication для тестов.
type MockApp struct {
	createEventFunc        func(params app.EventParams) error
	updateEventFunc        func(params app.EventParams) error
	deleteEventFunc        func(id string) error
	getEventByIDFunc       func(id string) (*storage.Event, error)
	listEventsForDayFunc   func(date time.Time) ([]storage.Event, error)
	listEventsForWeekFunc  func(startDate time.Time) ([]storage.Event, error)
	listEventsForMonthFunc func(startDate time.Time) ([]storage.Event, error)
}

func (m *MockApp) CreateEvent(params app.EventParams) error {
	if m.createEventFunc != nil {
		return m.createEventFunc(params)
	}
	return nil
}

func (m *MockApp) UpdateEvent(params app.EventParams) error {
	if m.updateEventFunc != nil {
		return m.updateEventFunc(params)
	}
	return nil
}

func (m *MockApp) DeleteEvent(id string) error {
	if m.deleteEventFunc != nil {
		return m.deleteEventFunc(id)
	}
	return nil
}

func (m *MockApp) GetEventByID(id string) (*storage.Event, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(id)
	}
	return nil, storage.ErrEventNotFound
}

func (m *MockApp) ListEventsForDay(date time.Time) ([]storage.Event, error) {
	if m.listEventsForDayFunc != nil {
		return m.listEventsForDayFunc(date)
	}
	return []storage.Event{}, nil
}

func (m *MockApp) ListEventsForWeek(startDate time.Time) ([]storage.Event, error) {
	if m.listEventsForWeekFunc != nil {
		return m.listEventsForWeekFunc(startDate)
	}
	return []storage.Event{}, nil
}

func (m *MockApp) ListEventsForMonth(startDate time.Time) ([]storage.Event, error) {
	if m.listEventsForMonthFunc != nil {
		return m.listEventsForMonthFunc(startDate)
	}
	return []storage.Event{}, nil
}

func TestCalendarHandlers_CreateEvent(t *testing.T) {
	logger := &MockLogger{}
	startTime := time.Date(2025, 12, 11, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(params app.EventParams) error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful creation",
			requestBody: api.CreateEventRequest{
				Title:     "Test Event",
				StartTime: startTime,
				Duration:  3600,
				UserId:    123,
			},
			mockFunc: func(params app.EventParams) error {
				return nil
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.EventResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Equal(t, "Test Event", resp.Title)
				require.Equal(t, int64(123), resp.UserId)
			},
		},
		{
			name: "validation error - empty title",
			requestBody: api.CreateEventRequest{
				Title:     "",
				StartTime: startTime,
				Duration:  3600,
				UserId:    123,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.ErrorResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Contains(t, resp.Error, "title is required")
			},
		},
		{
			name: "validation error - invalid duration",
			requestBody: api.CreateEventRequest{
				Title:     "Test",
				StartTime: startTime,
				Duration:  30, // меньше минимума (60)
				UserId:    123,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.ErrorResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Contains(t, resp.Error, "duration")
			},
		},
		{
			name: "date busy error",
			requestBody: api.CreateEventRequest{
				Title:     "Test Event",
				StartTime: startTime,
				Duration:  3600,
				UserId:    123,
			},
			mockFunc: func(params app.EventParams) error {
				return storage.ErrDateBusy
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.ErrorResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Equal(t, "date is busy", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				createEventFunc: tt.mockFunc,
				getEventByIDFunc: func(id string) (*storage.Event, error) {
					if tt.expectedStatus == http.StatusCreated {
						return &storage.Event{
							ID:        id,
							Title:     "Test Event",
							StartTime: startTime,
							Duration:  3600 * time.Second,
							UserID:    123,
						}, nil
					}
					return nil, storage.ErrEventNotFound
				},
			}

			handlers := NewCalendarHandlers(logger, mockApp)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.CreateEvent(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestCalendarHandlers_GetEvent(t *testing.T) {
	logger := &MockLogger{}
	startTime := time.Date(2025, 12, 11, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		eventID        string
		mockFunc       func(id string) (*storage.Event, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:    "successful get",
			eventID: testEventID,
			mockFunc: func(id string) (*storage.Event, error) {
				return &storage.Event{
					ID:        id,
					Title:     "Test Event",
					StartTime: startTime,
					Duration:  3600 * time.Second,
					UserID:    123,
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.EventResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Equal(t, "Test Event", resp.Title)
			},
		},
		{
			name:    "event not found",
			eventID: testEventID,
			mockFunc: func(id string) (*storage.Event, error) {
				return nil, storage.ErrEventNotFound
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.ErrorResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Equal(t, "event not found", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				getEventByIDFunc: tt.mockFunc,
			}

			handlers := NewCalendarHandlers(logger, mockApp)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+tt.eventID, nil)
			w := httptest.NewRecorder()

			handlers.GetEvent(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestCalendarHandlers_DeleteEvent(t *testing.T) {
	logger := &MockLogger{}

	tests := []struct {
		name           string
		eventID        string
		mockFunc       func(id string) error
		expectedStatus int
	}{
		{
			name:    "successful delete",
			eventID: testEventID,
			mockFunc: func(id string) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:    "event not found",
			eventID: testEventID,
			mockFunc: func(id string) error {
				return storage.ErrEventNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				deleteEventFunc: tt.mockFunc,
			}

			handlers := NewCalendarHandlers(logger, mockApp)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/"+tt.eventID, nil)
			w := httptest.NewRecorder()

			handlers.DeleteEvent(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCalendarHandlers_ListEventsForDay(t *testing.T) {
	logger := &MockLogger{}
	testDate := "2025-12-11"
	startTime := time.Date(2025, 12, 11, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		date           string
		mockFunc       func(date time.Time) ([]storage.Event, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful list",
			date: testDate,
			mockFunc: func(date time.Time) ([]storage.Event, error) {
				return []storage.Event{
					{
						ID:        "event-1",
						Title:     "Event 1",
						StartTime: startTime,
						Duration:  3600 * time.Second,
						UserID:    123,
					},
					{
						ID:        "event-2",
						Title:     "Event 2",
						StartTime: startTime.Add(2 * time.Hour),
						Duration:  1800 * time.Second,
						UserID:    123,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.EventListResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Len(t, resp.Events, 2)
				require.Equal(t, "Event 1", resp.Events[0].Title)
				require.Equal(t, "Event 2", resp.Events[1].Title)
			},
		},
		{
			name: "empty list",
			date: testDate,
			mockFunc: func(date time.Time) ([]storage.Event, error) {
				return []storage.Event{}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.EventListResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Len(t, resp.Events, 0)
			},
		},
		{
			name:           "invalid date format",
			date:           "invalid-date",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.ErrorResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Contains(t, resp.Error, "invalid date format")
			},
		},
		{
			name: "internal error",
			date: testDate,
			mockFunc: func(date time.Time) ([]storage.Event, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				listEventsForDayFunc: tt.mockFunc,
			}

			handlers := NewCalendarHandlers(logger, mockApp)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/events/day/"+tt.date, nil)
			w := httptest.NewRecorder()

			handlers.ListEventsForDay(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestCalendarHandlers_UpdateEvent(t *testing.T) {
	logger := &MockLogger{}
	startTime := time.Date(2025, 12, 11, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		eventID        string
		requestBody    interface{}
		mockUpdateFunc func(params app.EventParams) error
		mockGetFunc    func(id string) (*storage.Event, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:    "successful update",
			eventID: testEventID,
			requestBody: api.UpdateEventRequest{
				Title:     "Updated Event",
				StartTime: startTime,
				Duration:  7200,
				UserId:    123,
			},
			mockUpdateFunc: func(params app.EventParams) error {
				return nil
			},
			mockGetFunc: func(id string) (*storage.Event, error) {
				return &storage.Event{
					ID:        id,
					Title:     "Updated Event",
					StartTime: startTime,
					Duration:  7200 * time.Second,
					UserID:    123,
				}, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp api.EventResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				require.Equal(t, "Updated Event", resp.Title)
				require.Equal(t, int64(7200), resp.Duration)
			},
		},
		{
			name:    "event not found",
			eventID: testEventID,
			requestBody: api.UpdateEventRequest{
				Title:     "Updated Event",
				StartTime: startTime,
				Duration:  7200,
				UserId:    123,
			},
			mockUpdateFunc: func(params app.EventParams) error {
				return storage.ErrEventNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				updateEventFunc:  tt.mockUpdateFunc,
				getEventByIDFunc: tt.mockGetFunc,
			}

			handlers := NewCalendarHandlers(logger, mockApp)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+tt.eventID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.UpdateEvent(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}
