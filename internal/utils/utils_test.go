package utils

import (
	"robin2/internal/errors"
	"testing"
	"time"
)

var DateFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"02.01.2006 15:04:05",
	"02.01.2006 15:04",
	"02.01.2006T15:04:05Z",
	"02.01.2006T15:04:05",
	"02.01.2006T15:04",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05.0-07:00",
	"2006-01-02T15:04:05.00-07:00",
	"2006-01-02T15:04:05.000-07:00",
}

func Test_tryParseDate(t *testing.T) {
	// go through all date formats and check if they are valid
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
			err:      errors.ErrInvalidDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			date:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.ErrInvalidDate,
		},
	}

	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := TryParseDate(test.date, DateFormats)
			if err != test.err {
				t.Errorf("Test '%s' failed: expected error '%v', got '%v'", test.name, test.err, err)
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
			err:      errors.ErrInvalidDate,
		},
		{
			name:     "invalid 12.31.2022 00:00:00",
			time:     "12.31.2022 00:00:00",
			expected: time.Time{},
			err:      errors.ErrInvalidDate,
		},
	}

	for _, test := range test_cases {
		t.Run(test.name, func(t *testing.T) {
			date, err := ExcelTimeToTime(test.time, DateFormats)
			if err != test.err {
				t.Errorf("Test '%s' failed: expected error '%v', got '%v'", test.name, test.err, err)
			}
			if date != test.expected {
				t.Errorf("Test '%s' failed: expected '%v', got '%v'", test.name, test.expected, date)
			}

		})
	}
}
