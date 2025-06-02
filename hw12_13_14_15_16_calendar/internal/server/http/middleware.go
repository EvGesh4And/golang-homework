package internalhttp

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
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

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			// Дата и время.
			timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")

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

			logger.Info(fmt.Sprintf("%-16s [%s] %s %s %s %d %d %s",
				ip, timestamp, method, path, proto, status, latency, userAgent))
		})
	}
}
