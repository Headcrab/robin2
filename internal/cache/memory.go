package cache

import (
	"robin2/internal/config"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"sync"
	"time"
)

func init() {
	Register("memory", NewMemory)
}

type memoryTimeKey struct {
	tag  string
	date int64
}

type memoryStrKey struct {
	tag   string
	field string
}

type Memory struct {
	mu     sync.RWMutex
	cache  map[memoryTimeKey]float32
	fields map[memoryStrKey]float32
	config config.Config
}

func NewMemory(cfg config.Config) (Cache, error) {
	t := &Memory{
		cache:  make(map[memoryTimeKey]float32),
		fields: make(map[memoryStrKey]float32),
		config: cfg,
	}
	err := t.Connect()
	if err != nil {
		logger.Error(err.Error())
		return t, err
	}
	logger.Trace("NewMemoryCache")
	return t, nil
}

func (c *Memory) Connect() error {
	logger.Trace("cache connecting to memory")
	return nil
}

func (c *Memory) Disconnect() error {
	logger.Trace("cache disconnecting to memory")
	return nil
}

func (c *Memory) Get(tag string, date time.Time) (float32, error) {
	key := memoryTimeKey{tag: tag, date: date.Unix()}
	c.mu.RLock()
	t, ok := c.cache[key]
	c.mu.RUnlock()
	if !ok {
		return 0, errors.ErrKeyNotFound
	}
	return t, nil
}

func (c *Memory) Set(tag string, date time.Time, value float32) error {
	key := memoryTimeKey{tag: tag, date: date.Unix()}
	c.mu.Lock()
	c.cache[key] = value
	c.mu.Unlock()
	return nil
}

func (c *Memory) GetStr(tag string, field string) (float32, error) {
	key := memoryStrKey{tag: tag, field: field}
	c.mu.RLock()
	v, ok := c.fields[key]
	c.mu.RUnlock()
	if !ok {
		return 0, errors.ErrKeyNotFound
	}
	return v, nil
}

func (c *Memory) SetStr(tag string, field string, value float32) error {
	key := memoryStrKey{tag: tag, field: field}
	c.mu.Lock()
	c.fields[key] = value
	c.mu.Unlock()
	return nil
}
