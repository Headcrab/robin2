package format

import (
	"encoding/json"
	"fmt"
	"robin2/internal/data"
	"sort"
	"time"
)

// Регистрируем ResponseFormatterJSON при инициализации пакета
func init() {
	Register("json", &ResponseFormatterJSON{})
}

// ResponseFormatterJSON структура с полем округления
type ResponseFormatterJSON struct {
	round float64
}

// Process обрабатывает входные данные и возвращает отформатированную строку JSON в виде байтов
func (r *ResponseFormatterJSON) Process(val interface{}) []byte {
	var result interface{} // Интерфейс для хранения итоговой структуры данных перед сериализацией в JSON

	switch v := val.(type) {
	case float32:
		result = Round(v, r.round)

	case map[string]float32:
		processedMap := make(map[string]string)
		for k1, v1 := range v {
			processedMap[k1] = Format(Round(v1, r.round))
		}
		result = processedMap

	case map[string]map[time.Time]float32:
		processedMap := make(map[string]map[string]string)
		for k1, v1 := range v {
			innerMap := make(map[string]string)
			for k2, v2 := range v1 {
				innerMap[k2.Format("2006-01-02 15:04:05")] = Format(Round(v2, r.round))
			}
			processedMap[k1] = innerMap
		}
		result = processedMap

	case map[string]map[string]string:
		result = v

	case [][]string:
		result = v

	case []map[string]string:
		if len(v) > 0 {
			// Сортируем ключи
			keys := make([]string, 0, len(v[0]))
			for k1 := range v[0] {
				keys = append(keys, k1)
			}
			sort.Strings(keys)

			rows := make([]map[string]string, 0, len(v))
			for _, v1 := range v {
				row := make(map[string]string)
				for _, k1 := range keys {
					row[k1] = v1[k1]
				}
				rows = append(rows, row)
			}
			result = rows
		} else {
			result = v
		}

	case *data.Output:
		if len(v.Rows) == 1 {
			return []byte(fmt.Sprintf(`"%s"`, v.Rows[0][2]))
		}
		headers := make([]string, len(v.Headers))
		copy(headers, v.Headers)

		rows := make([][]string, len(v.Rows))
		for i, row := range v.Rows {
			rows[i] = make([]string, len(row))
			copy(rows[i], row)
		}
		result = map[string]interface{}{
			"headers": headers,
			"rows":    rows,
		}

	case []string:
		result = v

	case *data.Tag:
		result = map[string]interface{}{
			"value": Round(v.Value, r.round),
		}

	case data.Tags:
		tags := make([]map[string]interface{}, len(v))
		for i, tag := range v {
			tags[i] = map[string]interface{}{
				"name":  tag.Name,
				"date":  tag.Date.Format("2006-01-02 15:04:05"),
				"value": Round(tag.Value, r.round),
			}
		}
		result = tags

	default:
		// Если тип данных не поддерживается, возвращаем ошибку
		return []byte(fmt.Sprintf(`{"error":"ResponseFormatterJSON not supported: %v"}`, val))
	}

	// Преобразуем итоговую структуру в JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":"JSON marshaling error: %v"}`, err))
	}

	return jsonData
}

// SetRound устанавливает значение округления и возвращает сам объект
func (r *ResponseFormatterJSON) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}
