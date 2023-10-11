package timing

import (
	"bytes"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	rw.status = status
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriterWrapper{ResponseWriter: w, body: new(bytes.Buffer)}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		wrapper.Header().Set("X-Execution-Time", duration.String())

		if wrapper.status != 0 {
			wrapper.ResponseWriter.WriteHeader(wrapper.status)
		}
		wrapper.body.WriteTo(wrapper.ResponseWriter)
	})
}
