package cache

import (
	"fmt"
	"robin2/internal/config"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"sync"
)

var MemoryCacheByteLock = &sync.Mutex{}

type hash [16]byte

type MemoryCacheBytes interface {
	BaseCache
}

type MemoryCacheBytesImpl struct {
	MemoryCacheBytes
	cache  map[hash]float32
	config config.Config
}

func NewMemoryCacheByte(cfg config.Config) BaseCache {
	t := BaseCache(&MemoryCacheBytesImpl{
		cache:  make(map[hash]float32),
		config: cfg,
	})
	logger.Debug("NewMemoryCacheByte")
	return t
}

func (c *MemoryCacheBytesImpl) Connect() error {
	logger.Debug("cache connecting to memoryByte ")
	return nil
}

func (c *MemoryCacheBytesImpl) Disconnect() error {
	return nil
}

func (c *MemoryCacheBytesImpl) GetHash(key hash) (float32, error) {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	t := c.cache[key]
	if t == 0 {
		return 0, errors.ErrKeyNotFound
	}
	return t, nil
}

func (c *MemoryCacheBytesImpl) SetHash(key hash, value float32) error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	c.cache[key] = value
	logger.Debug("MemoryCacheByteImpl.Set size=" + fmt.Sprintf("%d", len(c.cache)))
	return nil
}

func (c *MemoryCacheBytesImpl) RemoveHash(key hash) error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	delete(c.cache, key)
	return nil
}

func (c *MemoryCacheBytesImpl) RemoveAll() error {
	MemoryCacheByteLock.Lock()
	defer MemoryCacheByteLock.Unlock()
	c.cache = make(map[hash]float32)
	return nil
}
