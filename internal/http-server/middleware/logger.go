package logger

import (
	"net/http"
	"time"

	"log/slog"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			next.ServeHTTP(w, r)

			log.Info(
				"mw info",
				slog.Any("method", r.Method),
				slog.Any("url path", r.URL.Path),
				slog.Any("remote addr", r.RemoteAddr),
				slog.Any("duration", time.Since(startTime).String()),
			)
		}

		return http.HandlerFunc(fn)
	}
}
