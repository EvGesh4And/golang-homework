package internalhttp

import (
	"fmt"
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

		// Захватываем IP клиента.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Оборачиваем writer, чтобы перехватить статус.
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Обработка запроса.
		next.ServeHTTP(recorder, r)

		// Метод, путь и версия.
		method := r.Method
		path := r.URL.RequestURI()
		proto := r.Proto

		// Статус.
		status := recorder.statusCode

		// Латентность.
		latency := time.Since(start).Milliseconds()

		// User-Agent.
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "-"
		} else {
			userAgent = `"` + userAgent + `"`
		}

		s.logger.Info(fmt.Sprintf("%s %s %s %s %d %d %s",
			ip, method, path, proto, status, latency, userAgent))
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
		s.logger.Debug("валидный Content-Type", "Content-Type", contentType)

		next.ServeHTTP(w, r)
	})
}
