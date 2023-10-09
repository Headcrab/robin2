package cache

import (
	"robin2/internal/config"
	"robin2/internal/logger"
)

type Factory interface {
	NewCache(string, config.Config) BaseCache
}

func NewFactory() Factory {
	logger.Debug("NewCacheFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewCache(cacheName string, cfg config.Config) BaseCache {
	switch cacheName {
	case "none":
		// logger.Info("no cache selected")
		return nil
	case "memory":
		// logger.Info("memory cache selected")
		return NewMemoryCache(cfg)
	case "memoryBytes":
		// logger.Info("memoryBytes cache selected")
		return NewMemoryCacheByte(cfg)
	case "redis":
		// logger.Info("redis cache selected")
		if c, err := NewRedisCache(cfg); err != nil {
			return nil
		} else {
			return c
		}
	default:
		logger.Info("no cache selected")
		return nil
	}
}
