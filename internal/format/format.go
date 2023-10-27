// todo: make formatter use map[string]map[time.Time]float32 and return one float32 if is
package format

import (
	"encoding/json"
	"encoding/xml"
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
	case "xml":
		return &ResponseFormatterXML{}
	case "raw":
		return &ResponseFormatterRaw{}
	default:
		return &ResponseFormatterString{}
	}
}

type ResponseFormatter interface {
	Process(val interface{}) []byte
	SetRound(r int) ResponseFormatter
}

func Round(val float32, round float64) float64 {
	return float64(math.Round(float64(val)*math.Pow(10, round)) / math.Pow(10, round))
}

func Format(val float64) string {
	return strings.Replace(strconv.FormatFloat(float64(val), 'f', -1, 64), ".", ",", -1)
}

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
	}
	return []byte("#Error: " + fmt.Sprint(val))
}

type ResponseFormatterString struct {
	round float64
}

func (r *ResponseFormatterString) Process(val interface{}) []byte {
	ret := ""
	switch v := val.(type) {
	case float32:
		ret = Format(Round(v, r.round))
	case map[string]float32:
		for k1, v1 := range v {
			ret += k1 + ";" + Format(Round(v1, r.round)) + "\n"
		}
	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + ";" + k2.Format("2006-01-02 15:04:05") + ";" + Format(Round(v2, r.round)) + "\n"
			}
		}
	case map[string]map[string]string:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + ";" + k2 + ";" + v2 + "\n"
			}
		}
	default:
		ret = "#Error: " + fmt.Sprint(val)
	}
	return []byte(fmt.Sprint(ret))
}

func (r *ResponseFormatterString) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		logger.Error(err.Error())
		return []byte("#Error: " + err.Error())
	}
	return data
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
		return mustMarshalXML(v)
	}
	return []byte("#Error: " + fmt.Sprint(val))
}

func mustMarshalXML(v interface{}) []byte {
	data, err := xml.MarshalIndent(v, "", " ")
	if err != nil {
		logger.Error(err.Error())
		return []byte("#Error: " + err.Error())
	}
	return data
}
