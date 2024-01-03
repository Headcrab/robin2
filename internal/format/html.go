package format

// todo: html, nothing

import (
	"fmt"
	"math"
	"time"
)

type ResponseFormatterHTML struct {
	round float64
}

func (r *ResponseFormatterHTML) Process(val interface{}) []byte {
	switch v := val.(type) {
	case float32:
		return []byte(fmt.Sprintf("%.2f", v))

	case map[string]float32:
		for k, v1 := range v {
			v[k] = float32(Round(v1, r.round))
		}

		return mustMarshalHTML(v)

	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				v[k1][k2] = float32(Round(v2, r.round))
			}
		}

		return mustMarshalHTML(v)

	case map[string]map[string]string:
		return mustMarshalHTML(v)
	}

	return []byte("ResponseFormatterHTML not supported:" + fmt.Sprintf("%v", val))
}

func NewResponseFormatterHTML(round float64) *ResponseFormatterHTML {
	return &ResponseFormatterHTML{
		round: round,
	}
}

func (r *ResponseFormatterHTML) SetRound(r2 int) ResponseFormatter {
	r.round = math.Pow(10, float64(r2))
	return r
}

func mustMarshalHTML(val interface{}) []byte {
	return []byte(fmt.Sprintf("%v", val))
}
