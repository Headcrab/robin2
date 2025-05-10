package format

import (
	"fmt"
	"robin2/internal/data"
	"sort"
	"strings"
	"time"
)

// Регистрируем ResponseFormatterString при инициализации пакета
func init() {
	Register("text", &ResponseFormatterString{})
}

// ResponseFormatterString структура с полем округления
type ResponseFormatterString struct {
	round float64
}

func (r *ResponseFormatterString) GetType() string { return "text" }

// Process обрабатывает входные данные и возвращает отформатированную строку в виде байтов
func (r *ResponseFormatterString) Process(val interface{}) []byte {
	var sb strings.Builder // Используем strings.Builder для эффективного построения строк

	// Обработка различных типов данных
	switch v := val.(type) {
	case float32:
		sb.WriteString(Format(Round(v, r.round)))

	case map[string]float32:
		for k1, v1 := range v {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", k1, Format(Round(v1, r.round))))
		}

	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", k1, k2.Format("2006-01-02 15:04:05"), Format(Round(v2, r.round))))
			}
		}

	case map[string]map[string]string:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", k1, k2, v2))
			}
		}

	case [][]string:
		for _, v1 := range v {
			sb.WriteString(strings.Join(v1, "\t"))
			sb.WriteString("\n")
		}

	case []map[string]string:
		if len(v) > 0 {
			// Сортируем ключи
			keys := make([]string, 0, len(v[0]))
			for k1 := range v[0] {
				keys = append(keys, k1)
			}
			sort.Strings(keys)
			sb.WriteString(strings.Join(keys, "\t"))
			sb.WriteString("\n")

			for _, v1 := range v {
				vals := make([]string, 0, len(keys))
				for _, k1 := range keys {
					vals = append(vals, v1[k1])
				}
				sb.WriteString(strings.Join(vals, "\t"))
				sb.WriteString("\n")
			}
		}

	case *data.Output:
		if len(v.Rows) == 1 {
			return []byte(v.Rows[0][0])
		}
		sb.WriteString(strings.Join(v.Headers, "\t"))
		sb.WriteString("\n")
		for _, row := range v.Rows {
			sb.WriteString(strings.Join(row, "\t"))
			sb.WriteString("\n")
		}

	case []string:
		sb.WriteString(strings.Join(v, "\n"))

	case *data.Tag:
		sb.WriteString(fmt.Sprintf("%v", Round(v.Value, r.round)))

	case data.Tags:
		for _, tag := range v {
			sb.WriteString(fmt.Sprintf("%v\t%v\t%v\n", tag.Name, tag.Date.Format("2006-01-02 15:04:05"), Round(tag.Value, r.round)))
		}

	default:
		// Возвращаем сообщение о неподдерживаемом типе
		sb.WriteString(fmt.Sprintf("ResponseFormatterString not supported: %v", val))
	}
	return []byte(sb.String())
}

// SetRound устанавливает значение округления и возвращает сам объект
func (r *ResponseFormatterString) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}
