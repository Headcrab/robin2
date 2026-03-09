package robin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"robin2/internal/utils"
	"runtime"
	"testing"
)

type tagListResponse struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// projectRoot вычисляет корень проекта из пути к файлу теста.
// internal/app -> ../.. -> корень
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// file: .../internal/app/app_test.go
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func Test_endpoint_get_tag_list(t *testing.T) {
	test_cases := []struct {
		name     string
		tag_like string
		expected string
	}{
		{
			name:     "valid endpoint",
			tag_like: "A20_WT_01%",
			expected: "A20_WT_01",
		},
		{
			name:     "invalid endpoint",
			tag_like: "/api/v1/tags/",
			expected: "",
		},
	}
	t.Setenv("TEST_WORK_DIR", projectRoot())
	app := NewApp()
	err := app.initDatabase()
	if err != nil {
		t.Skipf("пропущено: нет соединения с БД: %v", err)
	}
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			// test request, get response
			request := &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/get/tag/list/", RawQuery: "like=" + url.QueryEscape(test.tag_like)},
			}
			response := httptest.NewRecorder()
			app.handleAPIGetTagList(response, request)
			// check response
			if response.Code != http.StatusOK {
				t.Errorf("Test '%s' failed: expected status code '%v', got '%v'", test.name, http.StatusOK, response.Code)
			}
			// check response body
			var payload tagListResponse
			err := json.Unmarshal(response.Body.Bytes(), &payload)
			if err != nil {
				t.Errorf("Test '%s' failed: expected valid json, got '%v'", test.name, err)
			}
			if test.expected == "" {
				if len(payload.Rows) != 0 {
					t.Errorf("Test '%s' failed: expected 0 rows, got '%v'", test.name, len(payload.Rows))
				}
				return
			}

			if len(payload.Rows) == 0 {
				t.Fatalf("Test '%s' failed: expected at least 1 row, got 0", test.name)
			}

			found := false
			for _, row := range payload.Rows {
				if len(row) > 0 && row[0] == test.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Test '%s' failed: expected tag '%v' in rows '%v'", test.name, test.expected, payload.Rows)
			}
		})
	}
}

func Benchmark_NewApp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewApp()
	}
}
func Benchmark_excelTimeToTime(b *testing.B) {
	b.Setenv("TEST_WORK_DIR", projectRoot())
	app := NewApp()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = utils.TryParseDate("2019-01-01", app.config.DateFormats)
		if err != nil {
			fmt.Println(err)
		}
	}
}
func Benchmark_tryParseDate(b *testing.B) {
	b.Setenv("TEST_WORK_DIR", projectRoot())
	app := NewApp()
	// app := App{}
	for i := 0; i < b.N; i++ {
		_, _ = utils.TryParseDate("2019-01-01", app.config.DateFormats)
	}
}
