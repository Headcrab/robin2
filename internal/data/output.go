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
