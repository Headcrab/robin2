package cache

import (
	"robin2/internal/config"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"sync"
	"time"
)

var MemoryCacheLock = &sync.Mutex{}

type Memcache map[string]map[time.Time]float32

type Memory struct {
	// Cache
	cache  Memcache
	config config.Config
}

func NewMemory(cfg config.Config) (Memory, error) {
	t := Memory{
		cache:  make(Memcache),
		config: cfg,
	}
	t.Connect()
	logger.Trace("NewMemoryCache")
	return t, nil
}

func (c Memory) Connect() error {
	logger.Trace("cache connecting to memory")
	return nil
}

func (c Memory) Disconnect() error {
	logger.Trace("cache disconnecting to memory")
	return nil
}

func (c Memory) Get(tag string, date time.Time) (float32, error) {
	MemoryCacheLock.Lock()
	defer MemoryCacheLock.Unlock()
	t, ok := c.cache[tag][date]
	if !ok {
		return 0, errors.ErrKeyNotFound
	}
	return t, nil
}

func (c Memory) Set(tag string, date time.Time, value float32) error {
	MemoryCacheLock.Lock()
	defer MemoryCacheLock.Unlock()
	if t, ok := c.cache[tag]; ok {
		t[date] = value
	} else {
		c.cache[tag] = make(map[time.Time]float32)
		c.cache[tag][date] = value
	}
	return nil
}

func (c Memory) GetStr(tag string, field string) (float32, error) {
	v, err := c.Get(tag, time.Now())
	return v, err
}

func (c Memory) SetStr(tag string, field string, value float32) error {
	return c.Set(tag, time.Now(), value)
}
