package format

import (
	"encoding/json"
	"fmt"
	"math"
	"robin2/internal/logger"
	"strconv"
	"strings"
	"time"
)

func New(format string) ResponseFormatter {
	switch format {
	case "json":
		return &ResponseFormatterJSON{}
	case "raw":
		return &ResponseFormatterRaw{}
	default:
		return &ResponseFormatterString{}
	}
}

func Round[T float32 | float64](val T, round float64) float64 {
	return math.Round(float64(val)*math.Pow(10, round)) / math.Pow(10, round)
}

func Format[T float32 | float64](val T) string {
	return strings.Replace(strconv.FormatFloat(float64(val), 'f', -1, 64), ".", ",", -1)
}

func roundMap(data map[string]interface{}, round float64) map[string]interface{} {
	for k, v := range data {
		switch value := v.(type) {
		case float64:
			data[k] = Round(value, round)
		case []interface{}:
			roundSlice(value, round)
		case map[string]interface{}:
			roundMap(value, round)
		}
	}
	return data
}

func roundSlice(data []interface{}, round float64) []interface{} {
	for i, v := range data {
		switch value := v.(type) {
		case float32:
			data[i] = Round(value, round)
		case float64:
			data[i] = Round(value, round)
		case []interface{}:
			roundSlice(value, round)
		case map[string]interface{}:
			roundMap(value, round)
		}
	}
	return data
}

type ResponseFormatter interface {
	Process(val interface{}) []byte
	SetRound(r int) ResponseFormatter
}

type ResponseFormatterJSON struct {
	round float64
}

func mustMarshal(v interface{}) []byte {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		logger.Error(err.Error())
		return []byte("#Error: " + err.Error())
	}
	return data
}

func (r *ResponseFormatterJSON) Process(val interface{}) []byte {
	switch v := val.(type) {
	case float32:
		return mustMarshal(Round(float64(v), r.round))
	case float64:
		return mustMarshal(Round(v, r.round))
	case map[string]interface{}:
		return mustMarshal(roundMap(v, r.round))
	case []interface{}:
		return mustMarshal(roundSlice(v, r.round))
	}
	return mustMarshal(val)
}

func (r *ResponseFormatterJSON) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

type ResponseFormatterRaw struct {
	round float64
}

func (r *ResponseFormatterRaw) Process(val interface{}) []byte {
	return []byte(fmt.Sprintf("%v", val))
}

func (r *ResponseFormatterRaw) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

type ResponseFormatterString struct {
	round float64
}

func (r *ResponseFormatterString) Process(val interface{}) []byte {
	ret := ""
	switch v := val.(type) {
	case float64:
		ret = Format(Round(v, r.round))
	case float32:
		ret = Format(Round(v, r.round))
	case int64:
		ret = strconv.FormatInt(v, 10)
	case int32:
		ret = strconv.FormatInt(int64(v), 10)
	case int:
		ret = strconv.Itoa(v)
	case string:
		ret = v
	case []float32:
		vt := roundSlice(val.([]interface{}), r.round)
		for _, v := range vt {
			ret += fmt.Sprint(v) + "\n"
		}
	case []string:
		for _, v1 := range v {
			ret += fmt.Sprint(v1) + "\n"
		}
	case map[string]string:
		for k, v := range v {
			ret += k + ", " + v + "\n"
		}
	case map[string]float32:
		for k, v := range v {
			ret += k + ", " + Format(Round(v, r.round)) + "\n"
		}
	case map[time.Time]float32:
		for k, v := range v {
			ret += k.Format("2006-01-02 15:04:05") + ", " + Format(Round(v, r.round)) + "\n"
		}
	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + ", " + k2.Format("2006-01-02 15:04:05") + ", " + Format(Round(v2, r.round)) + "\n"
			}
		}
	case map[string]map[string]string:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + ", " + k2 + ", " + v2 + "\n"
			}
		}
	case []map[string]string:
		for _, v1 := range v {
			for _, v2 := range v1 {
				ret += v2 + "\t"
			}
			ret += "\n"
		}
	default:
		logger.Error("unknown type: " + fmt.Sprint(val))
	}
	return []byte(fmt.Sprint(ret))
}

func (r *ResponseFormatterString) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}
