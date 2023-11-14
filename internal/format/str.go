package format

import (
	"fmt"
	"robin2/internal/data"
	"sort"
	"strings"
	"time"
)

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
			ret += k1 + "\t" + Format(Round(v1, r.round)) + "\n"
		}

	case map[string]map[time.Time]float32:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + "\t" + k2.Format("2006-01-02 15:04:05") + "\t" + Format(Round(v2, r.round)) + "\n"
			}
		}

	case map[string]map[string]string:
		for k1, v1 := range v {
			for k2, v2 := range v1 {
				ret += k1 + "\t" + k2 + "\t" + v2 + "\n"
			}
		}

	case [][]string:
		ret = ""
		for _, v1 := range v {
			ret += strings.Join(v1, "\t") + "\n"
		}

	case []map[string]string:
		ret = ""
		// sort by key
		keys := make([]string, 0, len(v))

		for k1 := range v[0] {
			keys = append(keys, k1)
		}
		sort.Strings(keys)
		ret += strings.Join(keys, "\t") + "\n"

		for _, v1 := range v {
			t := []string{}
			for _, k1 := range keys {
				t = append(t, v1[k1])
			}
			ret += strings.Join(t, "\t") + "\n"
		}
		// if len(v) == 1 {
		// 	if len(v[0]) == 1 {
		// 		for _, v1 := range v[0] {
		// 			ret = v1
		// 		}
		// 	}
		// }
	case *data.Output:
		if len(v.Headers) == 1 && len(v.Rows) == 1 {
			return []byte(v.Rows[0][0])
		}
		s := ""
		for _, v1 := range v.Headers {
			s += v1 + "\t"
		}
		s += "\n"
		for _, v1 := range v.Rows {
			s += strings.Join(v1, "\t") + "\n"
		}
		ret = s

	case []string:
		ret = strings.Join(v, "\n")

	default:
		ret = "#Error: " + fmt.Sprint(val)
	}
	return []byte(fmt.Sprint(ret))
}

func (r *ResponseFormatterString) SetRound(r2 int) ResponseFormatter {
	r.round = float64(r2)
	return r
}
