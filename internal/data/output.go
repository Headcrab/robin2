package data

import "time"

type Output struct {
	Headers []string
	Rows    [][]string
	Count   int
	Err     error
}

type Tag struct {
	Name  string
	Date  time.Time
	Value float32
}

type Tags []Tag

func GetTags(tags map[string]map[time.Time]float32) Tags {
	t := make(Tags, 0, len(tags))
	for k, v := range tags {
		for k1, v1 := range v {
			t = append(t, Tag{Name: k, Date: k1, Value: v1})
		}
	}
	return t
}

func (t Tags) Len() int { return len(t) }

func (t Tags) Average() float32 {
	var sum float32
	for _, v := range t {
		sum += v.Value
	}
	return sum / float32(len(t))
}
