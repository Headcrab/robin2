package robin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsSameOriginLogClearRequestAllowsOriginHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Origin", "http://localhost:8008")

	if !isSameOriginLogClearRequest(req) {
		t.Fatal("expected same-origin request to be allowed")
	}
}

func TestIsSameOriginLogClearRequestAllowsRefererHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Referer", "http://localhost:8008/logs/")

	if !isSameOriginLogClearRequest(req) {
		t.Fatal("expected same-origin referer to be allowed")
	}
}

func TestIsSameOriginLogClearRequestRejectsForeignOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Origin", "http://evil.example")

	if isSameOriginLogClearRequest(req) {
		t.Fatal("expected foreign origin to be rejected")
	}
}
