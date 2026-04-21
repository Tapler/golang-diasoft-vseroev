package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	logger   Logger
	app      CalendarApplication
	handlers *CalendarHandlers
	server   *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

func NewServer(logger Logger, app CalendarApplication, host, port string) *Server {
	s := &Server{
		logger:   logger,
		app:      app,
		handlers: NewCalendarHandlers(logger, app),
	}

	mux := http.NewServeMux()

	// Регистрация API эндпоинтов
	mux.HandleFunc("/api/v1/events", s.eventsHandler)
	mux.HandleFunc("/api/v1/events/", s.eventsWithIDHandler)
	mux.HandleFunc("/api/v1/events/day/", s.handlers.ListEventsForDay)
	mux.HandleFunc("/api/v1/events/week/", s.handlers.ListEventsForWeek)
	mux.HandleFunc("/api/v1/events/month/", s.handlers.ListEventsForMonth)

	mux.HandleFunc("/hello", s.helloHandler)

	// Оборачиваем в middleware для логирования
	handler := loggingMiddleware(logger)(mux)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// eventsHandler обрабатывает POST /api/v1/events.
func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/events" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.handlers.CreateEvent(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// eventsWithIDHandler обрабатывает GET/PUT/DELETE /api/v1/events/{id}.
func (s *Server) eventsWithIDHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Проверяем, что это не эндпоинты day/week/month
	if strings.HasPrefix(path, "/api/v1/events/day/") ||
		strings.HasPrefix(path, "/api/v1/events/week/") ||
		strings.HasPrefix(path, "/api/v1/events/month/") {
		http.NotFound(w, r)
		return
	}

	// Проверяем, что есть ID после /api/v1/events/
	id := strings.TrimPrefix(path, "/api/v1/events/")
	if id == "" || id == "/" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handlers.GetEvent(w, r)
	case http.MethodPut:
		s.handlers.UpdateEvent(w, r)
	case http.MethodDelete:
		s.handlers.DeleteEvent(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(fmt.Sprintf("Starting HTTP server on %s", s.server.Addr))

	errChan := make(chan error, 1)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.logger.Info("Server stopped by context")
		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server...")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Handling /hello request from " + r.RemoteAddr)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello from Calendar!")); err != nil {
		s.logger.Error("Failed to write response: " + err.Error())
	}
}
