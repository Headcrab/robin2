package robin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"robin2/internal/errors"
	"robin2/internal/utils"
	"testing"
	"time"
)

func Test_tryParseDate(t *testing.T) {

	test_cases := []struct {
		name     string
		date     string
		expected time.Time
		err      error
	}{
		{
			name:     "valid 31.12.2022 00:00:00",
			date:     "31.12.2022 00:00:00",
			expected: time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
			err:      nil,
		},
		{
			name:     "valid 10.11.2022  18:12:34",
			date:     "10.11.2022  18:12:34",
			expected: time.Date(2022, 11, 10, 18, 12, 34, 0, time.Local),
			err:      nil,
		},
		{
			name:     "invalid empty string",
			date:     "",
			expected: time.Time{},
			err:      errors.InvalidDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			date:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.InvalidDate,
		},
	}
	app := NewApp()
	app.initDatabase()
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := utils.TryParseDate(test.date, app.config.GetStringSlice("app.date_formats"))
			if err != test.err {
				t.Errorf("Test '%s' failed: expected error '%v', got '%v'", test.name, test.err, err)
			}
			if date != test.expected {
				t.Errorf("Test '%s' failed: expected date '%v', got '%v'", test.name, test.expected, date)
			}
		})
	}
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
	app := NewApp()
	app.initDatabase()
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			// test request, get response
			request := &http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "/get/tag/list/?like=" + test.tag_like,
				},
			}
			response := httptest.NewRecorder()
			app.handleAPIGetTagList(response, request)
			// check response
			if response.Code != http.StatusOK {
				t.Errorf("Test '%s' failed: expected status code '%v', got '%v'", test.name, http.StatusOK, response.Code)
			}
			// check response body
			var tags []string
			err := json.Unmarshal(response.Body.Bytes(), &tags)
			if err != nil {
				t.Errorf("Test '%s' failed: expected valid json, got '%v'", test.name, err)
			}
			if len(tags) != 1 {
				t.Errorf("Test '%s' failed: expected 1 tag, got '%v'", test.name, len(tags))
			}
			if tags[0] != test.expected {
				t.Errorf("Test '%s' failed: expected tag '%v', got '%v'", test.name, test.expected, tags[0])
			}
		})
	}
}

func Test_excelTimeToTime(t *testing.T) {
	test_cases := []struct {
		name     string
		time     string
		expected time.Time
		err      error
	}{
		{
			name:     "valid 31.12.2022 00:00:00",
			time:     "44926,0",
			expected: time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
			err:      nil,
		},
		{
			name:     "valid 10.11.2022 18:12:34",
			time:     "44875.7587268519",
			expected: time.Date(2022, 11, 10, 18, 12, 34, 0, time.Local),
			err:      nil,
		},
		{
			name:     "invalid empty string",
			time:     "",
			expected: time.Time{},
			err:      errors.InvalidDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			time:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.InvalidDate,
		},
	}
	app := NewApp()
	// app.Init()
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := utils.ExcelTimeToTime(test.time, app.config.GetStringSlice("app.date_formats"))
			if err != test.err {
				t.Errorf("Test '%s' failed: expected error '%v', got '%v'", test.name, test.err, err)
			}
			if date != test.expected {
				t.Errorf("Test '%s' failed: expected '%v', got '%v'", test.name, test.expected, date)
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
	app := NewApp()
	for i := 0; i < b.N; i++ {
		utils.TryParseDate("2019-01-01", app.config.GetStringSlice("app.date_formats"))
	}
}
func Benchmark_tryParseDate(b *testing.B) {
	app := NewApp()
	// app := App{}
	for i := 0; i < b.N; i++ {
		utils.TryParseDate("2019-01-01", app.config.GetStringSlice("app.date_formats"))
	}
}
