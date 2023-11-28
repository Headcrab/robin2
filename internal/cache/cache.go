package cache

import (
	"robin2/internal/config"
	"robin2/internal/logger"
	"time"
)

type Cache interface {
	Connect() error
	Disconnect() error
	Get(tag string, date time.Time) (float32, error)
	Set(tag string, date time.Time, value float32) error
	GetStr(tag string, field string) (float32, error)
	SetStr(tag string, field string, value float32) error
}

func New(cfg config.Config) Cache {
	switch cfg.CurrCache.Type {
	case "memory":
		logger.Debug("NewCache.memory")
		t, err := NewMemory(cfg)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		return t
	case "memoryBytes":
		logger.Debug("NewCache.memoryBytes")
		t, err := NewMemoryByte(cfg)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		return t
	case "redis":
		logger.Debug("NewCache.redis")
		t, err := NewRedis(cfg)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		return t
	}
	logger.Error("cache type not found")
	return nil
}
