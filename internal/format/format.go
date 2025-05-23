// todo: make formatter use map[string]map[time.Time]float32 and return one float32 if is
package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

var (
	registry map[string]ResponseFormatter
	once     sync.Once
)

func Register(name string, format ResponseFormatter) {
	once.Do(func() {
		registry = make(map[string]ResponseFormatter)
	})
	registry[name] = format
}

func New(format string) (ResponseFormatter, error) {
	if format == "" {
		format = "text"
	}
	once.Do(func() {
		registry = make(map[string]ResponseFormatter)
	})
	formatter, ok := registry[format]
	if !ok {
		return nil, fmt.Errorf("formatter '%s' not found", format)
	}
	return formatter, nil
}

type ResponseFormatter interface {
	Process(val interface{}) []byte
	SetRound(r int) ResponseFormatter
	GetType() string
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

func (r *ResponseFormatterRaw) GetType() string {
	return "raw"
}

type FormatterPool struct {
	formatters chan ResponseFormatter
}

func NewFormatterPool(size int) *FormatterPool {
	return &FormatterPool{
		formatters: make(chan ResponseFormatter, size),
	}
}

func (p *FormatterPool) Get(format string) (ResponseFormatter, error) {
	var formatter ResponseFormatter
	var err error

	select {
	case f := <-p.formatters:
		if f.GetType() == format {
			return f, nil
		}
		// Если форматтер не того типа, который нам нужен,
		// вернем его в пул и создадим новый
		p.Put(f)
	default:
	}

	formatter, err = New(format)
	if err != nil {
		return nil, err
	}

	return formatter, nil
}

func (p *FormatterPool) Put(f ResponseFormatter) {
	select {
	case p.formatters <- f:
	default:
		// пул переполнен, пропускаем форматтер
	}
}
