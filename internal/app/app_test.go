package robin

import (
	"robin2/internal/errors"
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
			err:      errors.ErrEmptyDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			date:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.ErrInvalidDate,
		},
	}
	app := NewApp()
	app.init()
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := app.tryParseDate(test.date)
			if err != test.err {
				t.Errorf("Test '%s' failed: expected logger.Error '%v', got '%v'", test.name, test.err, err)
			}
			if date != test.expected {
				t.Errorf("Test '%s' failed: expected date '%v', got '%v'", test.name, test.expected, date)
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
			err:      errors.ErrEmptyDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			time:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.ErrInvalidDate,
		},
	}
	app := NewApp()
	// app.Init()
	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := app.excelTimeToTime(test.time)
			if err != test.err {
				t.Errorf("Test '%s' failed: expected logger.Error '%v', got '%v'", test.name, test.err, err)
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
		app.excelTimeToTime("2019-01-01")
	}
}
func Benchmark_tryParseDate(b *testing.B) {
	app := NewApp()
	// app := App{}
	for i := 0; i < b.N; i++ {
		app.tryParseDate("2019-01-01")
	}
}