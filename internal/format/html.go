package format

// todo: html, nothing

import (
	"fmt"
	"html"
	"robin2/internal/data"
	"robin2/internal/logger"
	"sort"
	"strings"
	"time"
)

func init() {
	Register("html", func() ResponseFormatter { return NewResponseFormatterHTML(2) })
}

type ResponseFormatterHTML struct {
	round float64
}

func (r *ResponseFormatterHTML) Process(val interface{}) []byte {
	switch v := val.(type) {
	case float32:
		return []byte(fmt.Sprintf("%.2f", v))

	case map[string]float32:
		return stringMapFloatToHTML(v, r.round)

	case map[string]map[time.Time]float32:
		return timedMapToHTML(v, r.round)

	case map[string]map[string]string:
		return nestedMapToHTML(v)

	case *data.Output:
		return outputToHTML(v)

	case []string:
		return stringListToHTML(v)

	case *data.Tag:
		return []byte(fmt.Sprintf("<pre>%s</pre>", html.EscapeString(Format(Round(v.Value, r.round)))))

	case data.Tags:
		return tagsToHTML(v, r.round)

	case logger.LogHistory:
		return logsToHTML([]logger.LogItem(v))

	case []logger.LogItem:
		return logsToHTML(v)
	}

	return []byte("ResponseFormatterHTML not supported:" + fmt.Sprintf("%v", val))
}

func NewResponseFormatterHTML(round float64) *ResponseFormatterHTML {
	return &ResponseFormatterHTML{
		round: round,
	}
}

func (r *ResponseFormatterHTML) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

func mustMarshalHTML(val interface{}) []byte {
	return []byte(fmt.Sprintf("%v", val))
}

func stringMapFloatToHTML(val map[string]float32, round float64) []byte {
	keys := make([]string, 0, len(val))
	for key := range val {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>")
	for _, key := range keys {
		sb.WriteString("<tr><td>")
		sb.WriteString(html.EscapeString(key))
		sb.WriteString("</td><td>")
		sb.WriteString(html.EscapeString(Format(Round(val[key], round))))
		sb.WriteString("</td></tr>")
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}

func timedMapToHTML(val map[string]map[time.Time]float32, round float64) []byte {
	keys := make([]string, 0, len(val))
	for key := range val {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Name</th><th>DateTime</th><th>Value</th></tr></thead><tbody>")
	for _, key := range keys {
		times := make([]time.Time, 0, len(val[key]))
		for ts := range val[key] {
			times = append(times, ts)
		}
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
		for _, ts := range times {
			sb.WriteString("<tr><td>")
			sb.WriteString(html.EscapeString(key))
			sb.WriteString("</td><td>")
			sb.WriteString(html.EscapeString(ts.Format("2006-01-02 15:04:05")))
			sb.WriteString("</td><td>")
			sb.WriteString(html.EscapeString(Format(Round(val[key][ts], round))))
			sb.WriteString("</td></tr>")
		}
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}

func nestedMapToHTML(val map[string]map[string]string) []byte {
	rowKeys := make([]string, 0, len(val))
	for key := range val {
		rowKeys = append(rowKeys, key)
	}
	sort.Strings(rowKeys)

	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Key</th><th>Field</th><th>Value</th></tr></thead><tbody>")
	for _, key := range rowKeys {
		fieldKeys := make([]string, 0, len(val[key]))
		for field := range val[key] {
			fieldKeys = append(fieldKeys, field)
		}
		sort.Strings(fieldKeys)
		for _, field := range fieldKeys {
			sb.WriteString("<tr><td>")
			sb.WriteString(html.EscapeString(key))
			sb.WriteString("</td><td>")
			sb.WriteString(html.EscapeString(field))
			sb.WriteString("</td><td>")
			sb.WriteString(html.EscapeString(val[key][field]))
			sb.WriteString("</td></tr>")
		}
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}

func outputToHTML(val *data.Output) []byte {
	if scalar, ok := scalarOutputValue(val); ok {
		return []byte("<pre>" + html.EscapeString(scalar) + "</pre>")
	}

	var sb strings.Builder
	sb.WriteString("<table><thead><tr>")
	for _, header := range val.Headers {
		sb.WriteString("<th>")
		sb.WriteString(html.EscapeString(header))
		sb.WriteString("</th>")
	}
	sb.WriteString("</tr></thead><tbody>")
	for _, row := range val.Rows {
		sb.WriteString("<tr>")
		for _, cell := range row {
			sb.WriteString("<td>")
			sb.WriteString(html.EscapeString(cell))
			sb.WriteString("</td>")
		}
		sb.WriteString("</tr>")
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}

func stringListToHTML(val []string) []byte {
	var sb strings.Builder
	sb.WriteString("<ul>")
	for _, item := range val {
		sb.WriteString("<li>")
		sb.WriteString(html.EscapeString(item))
		sb.WriteString("</li>")
	}
	sb.WriteString("</ul>")
	return []byte(sb.String())
}

func tagsToHTML(val data.Tags, round float64) []byte {
	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>Name</th><th>DateTime</th><th>Value</th></tr></thead><tbody>")
	for _, item := range val {
		sb.WriteString("<tr><td>")
		sb.WriteString(html.EscapeString(item.Name))
		sb.WriteString("</td><td>")
		sb.WriteString(html.EscapeString(item.Date.Format("2006-01-02 15:04:05")))
		sb.WriteString("</td><td>")
		sb.WriteString(html.EscapeString(Format(Round(item.Value, round))))
		sb.WriteString("</td></tr>")
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}

func logsToHTML(val []logger.LogItem) []byte {
	var sb strings.Builder
	sb.WriteString("<table><thead><tr><th>DateTime</th><th>Level</th><th>Message</th></tr></thead><tbody>")
	for _, item := range val {
		sb.WriteString("<tr><td>")
		sb.WriteString(html.EscapeString(item.Date.Format("2006-01-02 15:04:05")))
		sb.WriteString("</td><td>")
		sb.WriteString(html.EscapeString(item.Level))
		sb.WriteString("</td><td>")
		sb.WriteString(html.EscapeString(item.Msg))
		sb.WriteString("</td></tr>")
	}
	sb.WriteString("</tbody></table>")
	return []byte(sb.String())
}
