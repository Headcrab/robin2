package format

import (
	"encoding/json"
	"reflect"
	"robin2/internal/data"
	"robin2/internal/logger"
	"strings"
	"testing"
	"time"
)

func TestResponseFormatterJSONProcessOutputSingleColumn(t *testing.T) {
	fmtr := &ResponseFormatterJSON{}
	out := &data.Output{
		Headers: []string{"tag"},
		Rows:    [][]string{{"A20_WT_01"}},
	}

	raw := fmtr.Process(out)

	var got struct {
		Headers []string   `json:"headers"`
		Rows    [][]string `json:"rows"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", string(raw), err)
	}

	if !reflect.DeepEqual(got.Headers, out.Headers) {
		t.Fatalf("headers mismatch: got %v want %v", got.Headers, out.Headers)
	}
	if !reflect.DeepEqual(got.Rows, out.Rows) {
		t.Fatalf("rows mismatch: got %v want %v", got.Rows, out.Rows)
	}
}

func TestResponseFormatterStringProcessOutputSingleColumn(t *testing.T) {
	fmtr := &ResponseFormatterString{}
	out := &data.Output{
		Headers: []string{"tag"},
		Rows:    [][]string{{"A20_WT_01"}},
	}

	got := string(fmtr.Process(out))
	want := "tag\nA20_WT_01\n"
	if got != want {
		t.Fatalf("unexpected text output: got %q want %q", got, want)
	}
}

func TestResponseFormatterOutputScalarStillUsesValueColumn(t *testing.T) {
	out := &data.Output{
		Headers: []string{"tag", "date", "value"},
		Rows:    [][]string{{"A20_WT_01", "2026-03-13 08:41:17", "42"}},
	}

	jsonFormatter := &ResponseFormatterJSON{}
	if got := string(jsonFormatter.Process(out)); got != `"42"` {
		t.Fatalf("unexpected json scalar output: got %q", got)
	}

	stringFormatter := &ResponseFormatterString{}
	if got := string(stringFormatter.Process(out)); got != "42" {
		t.Fatalf("unexpected text scalar output: got %q", got)
	}
}

func TestResponseFormatterXMLProcessNestedStringMap(t *testing.T) {
	fmtr := &ResponseFormatterXML{}
	val := map[string]map[string]string{
		"A20_PMP01": {
			"tag_name":    "A20_PMP01",
			"description": "Pump & valve <main>",
		},
	}

	got := string(fmtr.Process(val))
	for _, want := range []string{
		"<data>",
		"<row>",
		"<key>A20_PMP01</key>",
		"<tag_name>A20_PMP01</tag_name>",
		"<description>Pump &amp; valve &lt;main&gt;</description>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected XML to contain %q, got %q", want, got)
		}
	}
	if strings.Contains(got, "#Error:") {
		t.Fatalf("unexpected XML formatter error: %q", got)
	}
}

func TestResponseFormatterHTMLProcessOutput(t *testing.T) {
	fmtr := &ResponseFormatterHTML{}
	out := &data.Output{
		Headers: []string{"tag", "value"},
		Rows:    [][]string{{"A20_PMP01", "42"}},
	}

	got := string(fmtr.Process(out))
	for _, want := range []string{"<table>", "<th>tag</th>", "<td>A20_PMP01</td>", "<td>42</td>"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected HTML to contain %q, got %q", want, got)
		}
	}
}

func TestResponseFormatterLogHistoryCompatibility(t *testing.T) {
	logs := logger.LogHistory{
		{
			Date:  time.Date(2026, 3, 13, 9, 30, 0, 0, time.UTC),
			Level: "INFO",
			Msg:   "started",
		},
	}

	xmlFormatter := &ResponseFormatterXML{}
	xml := string(xmlFormatter.Process(logs))
	if strings.Contains(xml, "not supported") || strings.Contains(xml, "#Error:") || !strings.Contains(xml, "<Message>started</Message>") {
		t.Fatalf("unexpected xml log output: %q", xml)
	}

	htmlFormatter := &ResponseFormatterHTML{}
	html := string(htmlFormatter.Process(logs))
	if strings.Contains(html, "not supported") || strings.Contains(html, "#Error:") || !strings.Contains(html, "<td>started</td>") {
		t.Fatalf("unexpected html log output: %q", html)
	}

	grafanaFormatter := &ResponseFormatterGrafana{}
	raw := grafanaFormatter.Process(logs)
	if strings.Contains(string(raw), "not supported") || strings.Contains(string(raw), "#Error:") {
		t.Fatalf("unexpected grafana log output: %q", string(raw))
	}
	var payload []map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected valid json for grafana logs, got %q: %v", string(raw), err)
	}
	if len(payload) != 1 || payload[0]["msg"] != "started" {
		t.Fatalf("unexpected grafana payload: %#v", payload)
	}
}
