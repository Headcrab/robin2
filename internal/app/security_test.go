package robin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSecureTestRequest(method, target string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("X-Admin-Token", "secret")
	return req, httptest.NewRecorder()
}

func TestTemplateHandlersRequireAdmin(t *testing.T) {
	app := &App{adminToken: "secret"}

	req := httptest.NewRequest(http.MethodGet, "/templ/list/", nil)
	w := httptest.NewRecorder()

	app.handleTemplateList(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without admin token, got %d", w.Code)
	}
}

func TestTemplateAddRequiresPost(t *testing.T) {
	app := &App{adminToken: "secret"}

	req := httptest.NewRequest(http.MethodGet, "/templ/add/", nil)
	req.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()

	app.handleTemplateAdd(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", w.Code)
	}
}

func TestTemplateListRejectsUnsafeMask(t *testing.T) {
	app := &App{adminToken: "secret"}
	req, w := newSecureTestRequest(http.MethodGet, "/templ/list/?like=%27%20OR%201=1--")

	app.handleTemplateList(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsafe mask, got %d", w.Code)
	}
}

func TestHandleAPIReloadConfigRequiresAdminAndPost(t *testing.T) {
	app := &App{adminToken: "secret"}

	req := httptest.NewRequest(http.MethodGet, "/api/reload/", nil)
	req.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()
	app.handleAPIReloadConfig(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET reload, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/reload/", nil)
	w = httptest.NewRecorder()
	app.handleAPIReloadConfig(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without admin token, got %d", w.Code)
	}
}

func TestIsSameOriginRequestAllowsOriginHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Origin", "http://localhost:8008")

	if !isSameOriginRequest(req) {
		t.Fatal("expected same-origin request to be allowed")
	}
}

func TestIsSameOriginRequestAllowsRefererHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Referer", "http://localhost:8008/logs/")

	if !isSameOriginRequest(req) {
		t.Fatal("expected same-origin referer to be allowed")
	}
}

func TestIsSameOriginRequestRejectsForeignOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8008/api/log/clear/", nil)
	req.Host = "localhost:8008"
	req.Header.Set("Origin", "http://evil.example")

	if isSameOriginRequest(req) {
		t.Fatal("expected foreign origin to be rejected")
	}
}
