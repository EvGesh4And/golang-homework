package internalhttp

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
)

// Обёртка для записи статуса ответа.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		s.logger.Info("http request finished",
			slog.String("ip", ip),
			slog.String("method", r.Method),
			slog.String("path", r.URL.RequestURI()),
			slog.String("proto", r.Proto),
			slog.Int("status", recorder.statusCode),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

func (s *Server) checkContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const requiredContentType = "application/json"
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, requiredContentType) {
			s.logger.Error(server.ErrInvalidContentType.Error(), "receivedContentType", contentType)
			http.Error(w, server.ErrInvalidContentType.Error(), http.StatusBadRequest)
			return
		}
		s.logger.Debug("valid Content-Type", "Content-Type", contentType)

		next.ServeHTTP(w, r)
	})
}
