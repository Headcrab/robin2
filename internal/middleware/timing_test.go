package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriterWrapperIgnoresDuplicateWriteHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wrapper := &responseWriterWrapper{ResponseWriter: rec}

	wrapper.WriteHeader(http.StatusCreated)
	wrapper.WriteHeader(http.StatusTeapot)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestResponseWriterWrapperWritesStatusOKByDefault(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	wrapper := &responseWriterWrapper{ResponseWriter: rec}

	if _, err := wrapper.Write([]byte("ok")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
