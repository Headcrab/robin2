package store

import (
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
