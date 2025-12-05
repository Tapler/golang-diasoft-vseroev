package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	logger Logger
	app    Application
	server *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

type Application interface {
	// Методы бизнес-логики будут добавлены позже
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	s := &Server{
		logger: logger,
		app:    app,
	}

	mux := http.NewServeMux()
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
