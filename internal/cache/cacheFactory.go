package cache

import (
	"robin2/pkg/logger"
)

type Factory interface {
	NewCache(string) BaseCache
}

func NewFactory() Factory {
	logger.Debug("NewCacheFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewCache(cacheName string) BaseCache {
	switch cacheName {
	case "none":
		// logger.Info("no cache selected")
		return nil
	case "memory":
		// logger.Info("memory cache selected")
		return NewMemoryCache()
	case "memoryBytes":
		// logger.Info("memoryBytes cache selected")
		return NewMemoryCacheByte()
	case "redis":
		// logger.Info("redis cache selected")
		return NewRedisCache()
	default:
		logger.Info("no cache selected")
		return nil
	}
}
