package format

import (
	"encoding/json"
	"fmt"
	"robin2/internal/data"
	"robin2/internal/logger"
	"time"
)

type ResponseFormatterJSON struct {
	round float64
}

func (r *ResponseFormatterJSON) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

func (r *ResponseFormatterJSON) Process(val interface{}) []byte {
	switch v := val.(type) {
	case float32:
		return mustMarshalJSON(Round(v, r.round))

	case map[string]float32:
		for k, v1 := range v {
			v[k] = float32(Round(v1, r.round))
		}
		return mustMarshalJSON(v)

	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				v[k1][k2] = float32(Round(v2, r.round))
			}
		}
		return mustMarshalJSON(v)

	case map[string]map[string]string:
		return mustMarshalJSON(v)

	case []string:
		return mustMarshalJSON(v)

	case *data.Output:
		if len(v.Headers) == 1 && len(v.Rows) == 1 {
			return []byte(v.Rows[0][0])
		}

		// return in format json
		var result []map[string]string
		for _, row := range v.Rows {
			item := make(map[string]string)
			for i, header := range v.Headers {
				item[header] = row[i]
			}
			result = append(result, item)
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			// handle error
			return nil
		}

		return jsonData
	}
	return []byte("#Error: " + fmt.Sprint(val))
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		logger.Error(err.Error())
		return []byte("#Error: " + err.Error())
	}
	return data
}