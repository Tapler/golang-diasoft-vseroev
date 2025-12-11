package internalhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/api"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
)

// CalendarHandlers содержит обработчики HTTP запросов для Calendar API.
type CalendarHandlers struct {
	logger Logger
	app    CalendarApplication
}

// CalendarApplication определяет интерфейс бизнес-логики календаря.
type CalendarApplication interface {
	CreateEvent(params app.EventParams) error
	UpdateEvent(params app.EventParams) error
	DeleteEvent(id string) error
	GetEventByID(id string) (*storage.Event, error)
	ListEventsForDay(date time.Time) ([]storage.Event, error)
	ListEventsForWeek(startDate time.Time) ([]storage.Event, error)
	ListEventsForMonth(startDate time.Time) ([]storage.Event, error)
}

// NewCalendarHandlers создает новый экземпляр обработчиков.
func NewCalendarHandlers(logger Logger, app CalendarApplication) *CalendarHandlers {
	return &CalendarHandlers{
		logger: logger,
		app:    app,
	}
}

// CreateEvent обрабатывает POST /api/v1/events.
func (h *CalendarHandlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateEventRequest
	if err := h.readJSON(r, &req); err != nil {
		h.logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validateCreateRequest(&req); err != nil {
		h.logger.Error(fmt.Sprintf("validation failed: %v", err))
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерация UUID для нового события
	id := generateUUID()

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	notifyBefore := time.Duration(0)
	if req.NotifyBefore != nil {
		notifyBefore = time.Duration(*req.NotifyBefore) * time.Second
	}

	// Вызов бизнес-логики
	err := h.app.CreateEvent(app.EventParams{
		ID:           id,
		Title:        req.Title,
		StartTime:    req.StartTime,
		Duration:     time.Duration(req.Duration) * time.Second,
		Description:  description,
		UserID:       req.UserId,
		NotifyBefore: notifyBefore,
	})
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	// Получаем созданное событие для ответа
	event, err := h.app.GetEventByID(id)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed to get created event: %v", err))
		h.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := h.eventToResponse(event)
	h.writeJSON(w, resp, http.StatusCreated)
}

// GetEvent обрабатывает GET /api/v1/events/{id}.
func (h *CalendarHandlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r.URL.Path, "/api/v1/events/")
	if id == "" {
		h.writeError(w, "missing event id", http.StatusBadRequest)
		return
	}

	event, err := h.app.GetEventByID(id)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	resp := h.eventToResponse(event)
	h.writeJSON(w, resp, http.StatusOK)
}

// UpdateEvent обрабатывает PUT /api/v1/events/{id}.
func (h *CalendarHandlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r.URL.Path, "/api/v1/events/")
	if id == "" {
		h.writeError(w, "missing event id", http.StatusBadRequest)
		return
	}

	var req api.UpdateEventRequest
	if err := h.readJSON(r, &req); err != nil {
		h.logger.Error(fmt.Sprintf("failed to decode request: %v", err))
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validateUpdateRequest(&req); err != nil {
		h.logger.Error(fmt.Sprintf("validation failed: %v", err))
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	notifyBefore := time.Duration(0)
	if req.NotifyBefore != nil {
		notifyBefore = time.Duration(*req.NotifyBefore) * time.Second
	}

	err := h.app.UpdateEvent(app.EventParams{
		ID:           id,
		Title:        req.Title,
		StartTime:    req.StartTime,
		Duration:     time.Duration(req.Duration) * time.Second,
		Description:  description,
		UserID:       req.UserId,
		NotifyBefore: notifyBefore,
	})
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	// Получаем обновленное событие для ответа
	event, err := h.app.GetEventByID(id)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed to get updated event: %v", err))
		h.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := h.eventToResponse(event)
	h.writeJSON(w, resp, http.StatusOK)
}

// DeleteEvent обрабатывает DELETE /api/v1/events/{id}.
func (h *CalendarHandlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r.URL.Path, "/api/v1/events/")
	if id == "" {
		h.writeError(w, "missing event id", http.StatusBadRequest)
		return
	}

	err := h.app.DeleteEvent(id)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEventsForDay обрабатывает GET /api/v1/events/day/{date}.
func (h *CalendarHandlers) ListEventsForDay(w http.ResponseWriter, r *http.Request) {
	h.listEventsHandler(w, r, "/api/v1/events/day/", h.app.ListEventsForDay)
}

// ListEventsForWeek обрабатывает GET /api/v1/events/week/{date}.
func (h *CalendarHandlers) ListEventsForWeek(w http.ResponseWriter, r *http.Request) {
	h.listEventsHandler(w, r, "/api/v1/events/week/", h.app.ListEventsForWeek)
}

// ListEventsForMonth обрабатывает GET /api/v1/events/month/{date}.
func (h *CalendarHandlers) ListEventsForMonth(w http.ResponseWriter, r *http.Request) {
	h.listEventsHandler(w, r, "/api/v1/events/month/", h.app.ListEventsForMonth)
}

// listEventsHandler общий обработчик для списков событий.
func (h *CalendarHandlers) listEventsHandler(
	w http.ResponseWriter,
	r *http.Request,
	pathPrefix string,
	listFunc func(time.Time) ([]storage.Event, error),
) {
	if r.Method != http.MethodGet {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dateStr := extractPathParam(r.URL.Path, pathPrefix)
	if dateStr == "" {
		h.writeError(w, "missing date", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Error(fmt.Sprintf("invalid date format: %v", err))
		h.writeError(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	events, err := listFunc(date)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed to list events: %v", err))
		h.writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := h.eventsToListResponse(events)
	h.writeJSON(w, resp, http.StatusOK)
}

// Вспомогательные методы

func (h *CalendarHandlers) readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return nil
}

func (h *CalendarHandlers) writeJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error(fmt.Sprintf("failed to encode response: %v", err))
	}
}

func (h *CalendarHandlers) writeError(w http.ResponseWriter, message string, statusCode int) {
	resp := api.ErrorResponse{
		Error: message,
	}
	h.writeJSON(w, resp, statusCode)
}

func (h *CalendarHandlers) handleBusinessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		h.writeError(w, "event not found", http.StatusNotFound)
	case errors.Is(err, storage.ErrDateBusy):
		h.writeError(w, "date is busy", http.StatusConflict)
	case errors.Is(err, storage.ErrInvalidEvent):
		h.writeError(w, "invalid event data", http.StatusBadRequest)
	default:
		h.logger.Error(fmt.Sprintf("business logic error: %v", err))
		h.writeError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *CalendarHandlers) validateCreateRequest(req *api.CreateEventRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if len(req.Title) > 200 {
		return errors.New("title too long (max 200 characters)")
	}
	if req.Duration < 60 || req.Duration > 86400 {
		return errors.New("duration must be between 60 and 86400 seconds")
	}
	if req.UserId < 1 {
		return errors.New("user_id must be positive")
	}
	if req.Description != nil && len(*req.Description) > 2000 {
		return errors.New("description too long (max 2000 characters)")
	}
	if req.NotifyBefore != nil && (*req.NotifyBefore < 0 || *req.NotifyBefore > 2592000) {
		return errors.New("notify_before must be between 0 and 2592000 seconds")
	}
	return nil
}

func (h *CalendarHandlers) validateUpdateRequest(req *api.UpdateEventRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if len(req.Title) > 200 {
		return errors.New("title too long (max 200 characters)")
	}
	if req.Duration < 60 || req.Duration > 86400 {
		return errors.New("duration must be between 60 and 86400 seconds")
	}
	if req.UserId < 1 {
		return errors.New("user_id must be positive")
	}
	if req.Description != nil && len(*req.Description) > 2000 {
		return errors.New("description too long (max 2000 characters)")
	}
	if req.NotifyBefore != nil && (*req.NotifyBefore < 0 || *req.NotifyBefore > 2592000) {
		return errors.New("notify_before must be between 0 and 2592000 seconds")
	}
	return nil
}

func (h *CalendarHandlers) eventToResponse(event *storage.Event) api.EventResponse {
	resp := api.EventResponse{
		Id:        parseUUID(event.ID),
		Title:     event.Title,
		StartTime: event.StartTime,
		Duration:  int64(event.Duration.Seconds()),
		UserId:    event.UserID,
	}

	if event.Description != "" {
		resp.Description = &event.Description
	}

	if event.NotifyBefore > 0 {
		notifyBefore := int64(event.NotifyBefore.Seconds())
		resp.NotifyBefore = &notifyBefore
	}

	return resp
}

func (h *CalendarHandlers) eventsToListResponse(events []storage.Event) api.EventListResponse {
	responses := make([]api.EventResponse, 0, len(events))
	for i := range events {
		responses = append(responses, h.eventToResponse(&events[i]))
	}

	return api.EventListResponse{
		Events: responses,
	}
}
