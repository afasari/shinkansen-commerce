package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", time.Since(start)),
			}

			if reqID, ok := r.Context().Value(RequestIDKey).(string); ok && reqID != "" {
				fields = append(fields, zap.String("request_id", reqID))
			}

			spanCtx := trace.SpanContextFromContext(r.Context())
			if spanCtx.HasTraceID() {
				fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()))
			}

			logger.Info("HTTP Request", fields...)
		})
	}
}
