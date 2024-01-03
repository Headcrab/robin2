package middleware

import (
	"net/http"
	"robin2/internal/logger"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
		next.ServeHTTP(w, r)
	})
}
