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

var cacheCreators = map[string]func(config.Config) BaseCache{
	"none": func(cfg config.Config) BaseCache {
		return nil
	},
	"memory": func(cfg config.Config) BaseCache {
		return NewMemoryCache(cfg)
	},
	"memoryBytes": func(cfg config.Config) BaseCache {
		return NewMemoryCacheByte(cfg)
	},
	"redis": func(cfg config.Config) BaseCache {
		if c, err := NewRedisCache(cfg); err != nil {
			return nil
		} else {
			return c
		}
	},
}

func (f *FactoryImpl) NewCache(cacheName string, cfg config.Config) BaseCache {
	if creator, ok := cacheCreators[cacheName]; ok {
		return creator(cfg)
	}
	logger.Info("no cache selected")
	return nil
}
