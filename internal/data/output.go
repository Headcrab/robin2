package data

import (
	"encoding/json"
	"time"
)

type Output struct {
	Headers []string
	Rows    [][]string
	Count   int
	Err     error
}

type Tag struct {
	Name  string    `json:"name"`
	Date  time.Time `json:"date"`
	Value float32   `json:"value"`
}

type Tags []*Tag

func GetTags(tags map[string]map[time.Time]float32) Tags {
	t := make(Tags, 0, len(tags))
	for k, v := range tags {
		for k1, v1 := range v {
			t = append(t, &Tag{Name: k, Date: k1, Value: v1})
		}
	}
	return t
}

func (t Tags) Len() int { return len(t) }

func (t Tags) Average(tag string) float32 {
	var sum float32
	for _, v := range t {
		if v.Name != tag {
			continue
		}
		sum += v.Value
	}
	return sum / float32(len(t))
}

func (t Tags) GetFromTo(from, to time.Time) Tags {
	tags := make(Tags, 0, len(t))
	for _, v := range t {
		if v.Date.After(from) && v.Date.Before(to) {
			tags = append(tags, v)
		}
	}
	return tags
}

// Функция для преобразования Tags в формат JSON для Grafana
func (tags Tags) ToGrafanaTimeSeries() ([]byte, error) {
	// Группировка данных по имени тега
	seriesMap := make(map[string][][2]interface{})
	for _, tag := range tags {
		if tag != nil {
			// Добавление точек данных в соответствующий временной ряд
			seriesMap[tag.Name] = append(seriesMap[tag.Name], [2]interface{}{
				tag.Value,
				tag.Date.UnixNano() / int64(time.Millisecond), // Преобразование времени в миллисекунды
			})
		}
	}

	// Создание слайса для JSON ответа
	var series []map[string]interface{}
	for name, datapoints := range seriesMap {
		series = append(series, map[string]interface{}{
			"target":     name,
			"datapoints": datapoints,
		})
	}

	// Конвертация в JSON
	return json.Marshal(series)
}

type Metric struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
}

type TimePoint struct {
	Time string   `json:"time"`
	Data []Metric `json:"data"`
}

func (tags Tags) ToCustomFormat() ([]byte, error) {
	// Словарь для хранения временных данных, ключом является строка времени
	timeDataMap := make(map[string][]Metric)

	for _, tag := range tags {
		// Форматируем время в строку согласно примеру
		timeStr := tag.Date.Format(time.RFC3339Nano)
		metric := Metric{
			Name:  tag.Name,
			Value: tag.Value,
		}
		timeDataMap[timeStr] = append(timeDataMap[timeStr], metric)
	}

	// Создаем результативный массив временных точек
	var result []TimePoint
	for timeStr, metrics := range timeDataMap {
		result = append(result, TimePoint{
			Time: timeStr,
			Data: metrics,
		})
	}

	return json.Marshal(result)
}
