package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (m *Middleware) WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggerResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(startTime)

		m.logger.Info(
			"",
			zap.String("URI", r.RequestURI),
			zap.String("METHOD", r.Method),
			zap.Duration("DURATION", duration),
			zap.Int("SIZE", responseData.size),
			zap.Int("STATUS", responseData.status),
		)
	})
}

func (r *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if err != nil {
		return size, fmt.Errorf("error write from loggerResponseWriter %w", err)
	}
	return size, nil
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
