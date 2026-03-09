package store

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestBuildBatchTagDateQuery(t *testing.T) {
	base := &Base{}
	date := time.Date(2026, 2, 19, 9, 0, 0, 0, time.UTC)

	t.Run("simple exact date query becomes IN", func(t *testing.T) {
		query, ok := base.buildBatchTagDateQuery(
			"select h.TagName, h.DateTime, h.Value from history h where (h.TagName) = '{tag}' and h.DateTime = '{date}'",
			[]string{"A20_WT_01", "A20_WT_02"},
			date,
		)
		if !ok {
			t.Fatalf("expected query to be buildable")
		}
		if !strings.Contains(query, "IN ('A20_WT_01','A20_WT_02')") {
			t.Fatalf("expected IN clause, got %q", query)
		}
		if !strings.Contains(query, "2026-02-19 09:00:00") {
			t.Fatalf("expected formatted date, got %q", query)
		}
	})

	t.Run("queries with order by stay on fallback path", func(t *testing.T) {
		_, ok := base.buildBatchTagDateQuery(
			"select h.TagName, h.DateTime, h.Value from history h where (h.TagName) = '{tag}' and h.DateTime <= '{date}' order by h.DateTime desc limit 1",
			[]string{"A20_WT_01", "A20_WT_02"},
			date,
		)
		if ok {
			t.Fatalf("expected fallback for non-batch-safe query")
		}
	})
}

func TestSampleTagCountValues(t *testing.T) {
	from := time.Date(2023, 1, 30, 17, 10, 0, 0, time.UTC)

	t.Run("holds last known value on gaps", func(t *testing.T) {
		values := map[string]float32{
			from.Format("2006-01-02 15:04:05"):                   15.12,
			from.Add(8 * time.Second).Format("2006-01-02 15:04:05"): 15.17,
		}

		got, err := sampleTagCountValues(from, 9, 1, func(ts time.Time) (float32, error) {
			if val, ok := values[ts.Format("2006-01-02 15:04:05")]; ok {
				return val, nil
			}
			return 0, sql.ErrNoRows
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got[from.Add(1*time.Second)] != 15.12 {
			t.Fatalf("expected carry-forward value 15.12 at +1s, got %v", got[from.Add(1*time.Second)])
		}
		if got[from.Add(7*time.Second)] != 15.12 {
			t.Fatalf("expected carry-forward value 15.12 at +7s, got %v", got[from.Add(7*time.Second)])
		}
		if got[from.Add(8*time.Second)] != 15.17 {
			t.Fatalf("expected fresh value 15.17 at +8s, got %v", got[from.Add(8*time.Second)])
		}
	})

	t.Run("leading gaps become -1 until first value appears", func(t *testing.T) {
		got, err := sampleTagCountValues(from, 3, 1, func(ts time.Time) (float32, error) {
			if ts.Equal(from.Add(2 * time.Second)) {
				return 42, nil
			}
			return 0, sql.ErrNoRows
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got[from] != -1 {
			t.Fatalf("expected -1 for first missing point, got %v", got[from])
		}
		if got[from.Add(1*time.Second)] != -1 {
			t.Fatalf("expected -1 for second missing point, got %v", got[from.Add(1*time.Second)])
		}
		if got[from.Add(2*time.Second)] != 42 {
			t.Fatalf("expected 42 for first real point, got %v", got[from.Add(2*time.Second)])
		}
	})
}
