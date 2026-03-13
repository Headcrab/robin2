package format

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"robin2/internal/data"
	"robin2/internal/logger"
	"sort"
	"strings"
	"time"
)

func init() {
	Register("xml", func() ResponseFormatter { return &ResponseFormatterXML{} })
}

type ResponseFormatterXML struct {
	round float64
}

func (r *ResponseFormatterXML) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

func (r *ResponseFormatterXML) Process(val interface{}) []byte {
	switch v := val.(type) {
	case float32:
		return mustMarshalXML(Round(v, r.round))

	case map[string]float32:
		for k, v1 := range v {
			v[k] = float32(Round(v1, r.round))
		}
		return mustMarshalXML(v)

	case map[string]map[time.Time]float32:
		s := "<data>\n"
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				s += "\t<row>\n"
				s += "\t\t<TagName>" + k1 + "</TagName>\n"
				s += "\t\t<DateTime>" + k2.Format("2006-01-02 15:04:05") + "</DateTime>\n"
				s += "\t\t<Value>" + fmt.Sprintf("%v", Format(Round(v2, r.round))) + "</Value>\n"
				s += "\t</row>\n"
			}
		}
		s += "</data>"
		return []byte(s)

	case map[string]map[string]string:
		return nestedStringMapToXML(v)

	case []map[string]string:
		s := "<data>\n"
		for _, v1 := range v {
			s += "\t<row>\n"
			for k2, v2 := range v1 {
				s += "\t\t<" + k2 + ">" + v2 + "</" + k2 + ">\n"
			}
			s += "\t</row>\n"
		}
		s += "</data>"
		return []byte(s)

	case [][]string:
		s := "<data>\n"
		for _, v1 := range v {
			s += "\t<row>\n"
			for _, v2 := range v1 {
				s += "\t\t<" + v2 + "></" + v2 + ">\n"
			}
			s += "\t</row>\n"
		}
		s += "</data>"
		return []byte(s)

	case []string:
		s := "<data>\n"
		for _, v1 := range v {
			s += "\t<row>\n" + "\t\t<" + v1 + "></" + v1 + ">\n" + "\t</row>\n"
		}
		s += "</data>"
		return []byte(s)

	case *data.Output:
		if len(v.Headers) == 1 && len(v.Rows) == 1 {
			return []byte(v.Rows[0][0])
		}
		s := "<data>\n"
		for _, v1 := range v.Rows {
			s += "\t<row>\n"
			for k2, v2 := range v1 {
				s += "\t\t<" + v.Headers[k2] + ">" + v2 + "</" + v.Headers[k2] + ">\n"
			}
			s += "\t</row>\n"
		}
		s += "</data>"
		return []byte(s)
	case data.Tags:
		s := "<data>\n"
		for _, v1 := range v {
			s += "\t<row>\n"
			s += "\t\t<TagName>" + v1.Name + "</TagName>\n"
			s += "\t\t<DateTime>" + v1.Date.Format("2006-01-02 15:04:05") + "</DateTime>\n"
			s += "\t\t<Value>" + fmt.Sprintf("%v", Format(Round(v1.Value, r.round))) + "</Value>\n"
			s += "\t</row>\n"
		}
		s += "</data>"
		return []byte(s)
	case logger.LogHistory:
		return logItemsToXML([]logger.LogItem(v))
	case []logger.LogItem:
		return logItemsToXML(v)
	}
	return []byte("ResponseFormatterXML not supported:" + fmt.Sprint(val))
}

func nestedStringMapToXML(val map[string]map[string]string) []byte {
	var sb strings.Builder
	sb.WriteString("<data>\n")

	rowKeys := make([]string, 0, len(val))
	for key := range val {
		rowKeys = append(rowKeys, key)
	}
	sort.Strings(rowKeys)

	for _, rowKey := range rowKeys {
		sb.WriteString("\t<row>\n")
		sb.WriteString("\t\t<key>")
		sb.WriteString(escapeXMLText(rowKey))
		sb.WriteString("</key>\n")

		fieldKeys := make([]string, 0, len(val[rowKey]))
		for field := range val[rowKey] {
			fieldKeys = append(fieldKeys, field)
		}
		sort.Strings(fieldKeys)

		for _, field := range fieldKeys {
			sb.WriteString("\t\t<")
			sb.WriteString(field)
			sb.WriteString(">")
			sb.WriteString(escapeXMLText(val[rowKey][field]))
			sb.WriteString("</")
			sb.WriteString(field)
			sb.WriteString(">\n")
		}

		sb.WriteString("\t</row>\n")
	}

	sb.WriteString("</data>")
	return []byte(sb.String())
}

func escapeXMLText(val string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(val)); err != nil {
		logger.Error(err.Error())
		return val
	}
	return buf.String()
}

func logItemsToXML(items []logger.LogItem) []byte {
	var sb strings.Builder
	sb.WriteString("<data>\n")
	for _, item := range items {
		sb.WriteString("\t<row>\n")
		sb.WriteString("\t\t<DateTime>")
		sb.WriteString(escapeXMLText(item.Date.Format("2006-01-02 15:04:05")))
		sb.WriteString("</DateTime>\n")
		sb.WriteString("\t\t<Level>")
		sb.WriteString(escapeXMLText(item.Level))
		sb.WriteString("</Level>\n")
		sb.WriteString("\t\t<Message>")
		sb.WriteString(escapeXMLText(item.Msg))
		sb.WriteString("</Message>\n")
		sb.WriteString("\t</row>\n")
	}
	sb.WriteString("</data>")
	return []byte(sb.String())
}

func mustMarshalXML(v interface{}) []byte {
	data, err := xml.MarshalIndent(v, "", " ")
	if err != nil {
		logger.Error(err.Error())
		return []byte("#Error: " + err.Error())
	}
	return data
}
