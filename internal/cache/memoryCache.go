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

type MemoryCache interface {
	BaseCache
}

type MemoryCacheImpl struct {
	MemoryCache
	cache  Memcache
	config config.Config
}

func NewMemoryCache(cfg config.Config) BaseCache {
	t := BaseCache(&MemoryCacheImpl{
		cache:  make(Memcache),
		config: cfg,
	})
	t.Connect()
	logger.Trace("NewMemoryCache")
	return t
}

func (c *MemoryCacheImpl) Connect() error {
	logger.Trace("cache connected to memory")
	return nil
}

func (c *MemoryCacheImpl) Get(tag string, date time.Time) (float32, error) {
	MemoryCacheLock.Lock()
	defer MemoryCacheLock.Unlock()
	t, ok := c.cache[tag][date]
	if !ok {
		return 0, errors.KeyNotFound
	}
	return t, nil
}

func (c *MemoryCacheImpl) Set(tag string, date time.Time, value float32) error {
	MemoryCacheLock.Lock()
	defer MemoryCacheLock.Unlock()
	if t, ok := c.cache[tag]; ok {
		t[date] = value
	} else {
		c.cache[tag] = make(map[time.Time]float32)
		c.cache[tag][date] = value
	}
	// logger.Trace(fmt.Sprintf("MemoryCacheImpl.Set tag[%d][%d] ", len(c.cache), len(c.cache[tag])))
	return nil
}

// func (c *MemoryCacheImpl) Remove(key string) error {
// 	MemoryCacheLock.Lock()
// 	defer MemoryCacheLock.Unlock()
// 	delete(c.cache, key)
// 	return nil
// }

// func (c *MemoryCacheImpl) RemoveAll() error {
// 	MemoryCacheLock.Lock()
// 	defer MemoryCacheLock.Unlock()
// 	logger.Debug("MemoryCacheImpl.RemoveAll size="+fmt.Sprintf("%d", len(c.cache)))
// 	c.cache = make(map[string]string)
// 	return nil
// }
