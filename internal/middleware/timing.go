package middleware

import (
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	if rw.wrote {
		return
	}
	rw.wrote = true
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wrote {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func Timing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Add("Trailer", "X-Execution-Time")
		wrapper := &responseWriterWrapper{ResponseWriter: w}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		w.Header().Set("X-Execution-Time", duration.String())
	})
}
