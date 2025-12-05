package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

// responseWriter оборачивает http.ResponseWriter для захвата статус-кода.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Оборачиваем ResponseWriter для захвата статус-кода
			rw := newResponseWriter(w)

			// Обрабатываем запрос
			next.ServeHTTP(rw, r)

			// Вычисляем latency
			latency := time.Since(start)

			// Получаем IP клиента
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			}

			// Получаем User-Agent
			userAgent := r.Header.Get("User-Agent")
			if userAgent == "" {
				userAgent = "-"
			}

			// Формируем лог в формате:
			// IP [DateTime] Method Path Protocol StatusCode Latency "UserAgent"
			logMsg := fmt.Sprintf("%s [%s] %s %s %s %d %v \"%s\"",
				clientIP,
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.Path,
				r.Proto,
				rw.statusCode,
				latency,
				userAgent,
			)

			logger.Info(logMsg)
		})
	}
}
