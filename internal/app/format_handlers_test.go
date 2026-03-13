package robin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/data"
	"robin2/internal/format"
	"strings"
	"testing"
	"time"
)

type formatTestStore struct{}

func (s *formatTestStore) Connect(name string, cache cache.Cache) error { return nil }
func (s *formatTestStore) GetTagDate(tag string, date time.Time) (*data.Tag, error) {
	return &data.Tag{Name: tag, Date: date, Value: 42.5}, nil
}
func (s *formatTestStore) GetTagsDate(tags []string, date time.Time) (data.Tags, error) {
	return nil, nil
}
func (s *formatTestStore) GetTagCount(tag string, from time.Time, to time.Time, strCount int) (map[string]map[time.Time]float32, error) {
	return nil, nil
}
func (s *formatTestStore) GetTagCountGroup(tag string, from time.Time, to time.Time, strCount int, group string) (data.Tags, error) {
	return nil, nil
}
func (s *formatTestStore) GetTagFromTo(tag string, from time.Time, to time.Time) (data.Tags, error) {
	return nil, nil
}
func (s *formatTestStore) GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error) {
	return 0, nil
}
func (s *formatTestStore) GetTagList(like string) (*data.Output, error) {
	return &data.Output{Headers: []string{"tag"}, Rows: [][]string{{"A20_PMP01"}}}, nil
}
func (s *formatTestStore) GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	return nil, nil
}
func (s *formatTestStore) GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	return nil, nil
}
func (s *formatTestStore) GetStatus() (string, time.Duration, error) { return "ok", time.Minute, nil }
func (s *formatTestStore) TemplateList(like string) (map[string]string, error) {
	return map[string]string{"demo": "select 1"}, nil
}
func (s *formatTestStore) TemplateExec(name string, params map[string]string) (*data.Output, error) {
	return &data.Output{Headers: []string{"tag", "date", "value"}, Rows: [][]string{{"A20_PMP01", "2026-03-13 09:00:00", "42"}}}, nil
}
func (s *formatTestStore) TemplateAdd(name string, body string) error { return nil }
func (s *formatTestStore) TemplateSet(name string, body string) error { return nil }
func (s *formatTestStore) TemplateGet(name string) (string, error)    { return "select 1", nil }
func (s *formatTestStore) TemplateDel(name string) error              { return nil }
func (s *formatTestStore) ExecQuery(query string) (*data.Output, error) {
	return &data.Output{Headers: []string{"value"}, Rows: [][]string{{"42"}}}, nil
}

func newFormatTestApp() *App {
	return &App{
		adminToken: "secret",
		config: config.Config{
			Round:       2,
			DateFormats: []string{"2006-01-02 15:04:05"},
		},
		store:         &formatTestStore{},
		formatterPool: format.NewFormatterPool(2),
	}
}

func TestHandleAPIGetTagSetsContentTypeByFormat(t *testing.T) {
	app := newFormatTestApp()
	req := httptest.NewRequest(http.MethodGet, "/get/tag/?tag=A20_PMP01&date=2026-03-13%2009:00:00&format=json", nil)
	w := httptest.NewRecorder()

	app.handleAPIGetTag(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	if w.Body.String() != `{"value":42.5}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleTemplateExecDefaultsToTextFormat(t *testing.T) {
	app := newFormatTestApp()
	form := url.Values{}
	form.Set("name", "demo")

	req := httptest.NewRequest(http.MethodPost, "/templ/exec/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()

	app.handleTemplateExec(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected text/plain; charset=utf-8, got %q", ct)
	}
	if w.Body.String() != "42" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
