// todo: make formatter use map[string]map[time.Time]float32 and return one float32 if is
package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func New(format string) ResponseFormatter {
	switch format {
	case "json":
		return &ResponseFormatterJSON{}

	case "xml":
		return &ResponseFormatterXML{}

	case "raw":
		return &ResponseFormatterRaw{}

	case "html":
		return &ResponseFormatterHTML{}

	case "str":
		return &ResponseFormatterString{}
	default:
		return &ResponseFormatterString{}
	}
	// return nil
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
