// todo: make formatter use map[string]map[time.Time]float32 and return one float32 if is
package format

import (
	"fmt"
	"math"
	"robin2/internal/data"
	"strconv"
	"strings"
	"sync"
)

var (
	registry   = make(map[string]func() ResponseFormatter)
	registryMu sync.RWMutex
)

func Register(name string, factory func() ResponseFormatter) {
	registryMu.Lock()
	registry[name] = factory
	registryMu.Unlock()
}

func New(format string) (ResponseFormatter, error) {
	registryMu.RLock()
	factory, ok := registry[format]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("formatter '%s' not found", format)
	}
	formatter := factory()
	if formatter == nil {
		return nil, fmt.Errorf("formatter '%s' factory returned nil", format)
	}
	return formatter, nil
}

type ResponseFormatter interface {
	Process(val interface{}) []byte
	SetRound(r int) ResponseFormatter
}

func scalarOutputValue(out *data.Output) (string, bool) {
	rows := out.Rows
	if len(rows) != 1 {
		return "", false
	}

	row := rows[0]
	if len(row) >= 3 {
		return row[2], true
	}

	return "", false
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

type FormatterPool struct {
	size       int
	mu         sync.Mutex
	formatters map[string]chan ResponseFormatter
}

func NewFormatterPool(size int) *FormatterPool {
	if size < 1 {
		size = 1
	}
	return &FormatterPool{
		size:       size,
		formatters: make(map[string]chan ResponseFormatter),
	}
}

func (p *FormatterPool) Get(format string) (ResponseFormatter, error) {
	pool := p.getPool(format)
	select {
	case f := <-pool:
		return f, nil
	default:
		return New(format)
	}
}

func (p *FormatterPool) Put(format string, f ResponseFormatter) {
	if f == nil {
		return
	}
	pool := p.getPool(format)
	select {
	case pool <- f:
	default:
		// пул переполнен, пропускаем форматтер
	}
}

func (p *FormatterPool) getPool(format string) chan ResponseFormatter {
	p.mu.Lock()
	defer p.mu.Unlock()

	pool, ok := p.formatters[format]
	if !ok {
		pool = make(chan ResponseFormatter, p.size)
		p.formatters[format] = pool
	}
	return pool
}
