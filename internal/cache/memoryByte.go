package cache

import (
	"fmt"
	"robin2/internal/config"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"sync"
	"time"
)

var MemoryCacheByteLock = &sync.Mutex{}

type hash [16]byte

type MemoryByte struct {
	// Cache
	cache  map[hash]float32
	config config.Config
}

func NewMemoryByte(cfg config.Config) (*MemoryByte, error) {
	t := MemoryByte{
		cache:  make(map[hash]float32),
		config: cfg,
	}
	logger.Debug("NewMemoryCacheByte")
	return &t, nil
}

func (c MemoryByte) Connect() error {
	logger.Debug("cache connecting to memoryByte ")
	return nil
}

func (c MemoryByte) Disconnect() error {
	return nil
}

func (c MemoryByte) GetHash(key hash) (float32, error) {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	t := c.cache[key]
	if t == 0 {
		return 0, errors.ErrKeyNotFound
	}
	return t, nil
}

func (c MemoryByte) SetHash(key hash, value float32) error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	c.cache[key] = value
	logger.Debug("MemoryCacheByteImpl.Set size=" + fmt.Sprintf("%d", len(c.cache)))
	return nil
}

func (c MemoryByte) RemoveHash(key hash) error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	delete(c.cache, key)
	return nil
}

func (c *MemoryByte) RemoveAll() error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	c.cache = make(map[hash]float32)
	return nil
}

func (c MemoryByte) Get(tag string, date time.Time) (float32, error) {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	t, ok := c.cache[hash([]byte(tag+date.Format("2006-01-02 15:04:05")))]
	if !ok {
		return 0, errors.ErrKeyNotFound
	}
	return t, nil
}

func (c MemoryByte) Set(tag string, date time.Time, value float32) error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	c.cache[hash([]byte(tag+date.Format("2006-01-02 15:04:05")))] = value
	return nil
}

func (c MemoryByte) GetStr(tag string, field string) (float32, error) {
	logger.Trace("MemoryCacheByteImpl.GetStr unimplemented")
	return 0, nil
}

func (c MemoryByte) SetStr(tag string, field string, value float32) error {
	logger.Trace("MemoryCacheByteImpl.SetStr unimplemented")
	return nil
}
