package cache

import (
	"robin2/pkg/logger"
)

type Factory interface {
	NewCache(string) *BaseCache
}

func NewCacheFactory() Factory {
	logger.Log(logger.Debug, "NewCacheFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewCache(cacheName string) *BaseCache {
	switch cacheName {
	case "memory":
		// logger.Log(logger.Info, "memory cache selected")
		return NewMemoryCache()
	case "memoryBytes":
		// logger.Log(logger.Info, "memoryBytes cache selected")
		return NewMemoryCacheByte()
	case "redis":
		// logger.Log(logger.Info, "redis cache selected")
		return NewRedisCache()
	default:
		logger.Log(logger.Info, "no cache selected")
		return nil
	}
}
