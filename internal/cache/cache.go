package cache

import (
	"robin2/internal/config"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"time"
)

var registry map[string]func(config.Config) (Cache, error)

func Register(name string, f func(config.Config) (Cache, error)) {
	if registry == nil {
		registry = make(map[string]func(config.Config) (Cache, error))
	}
	registry[name] = f
}

type Cache interface {
	Connect() error
	Disconnect() error
	Get(tag string, date time.Time) (float32, error)
	Set(tag string, date time.Time, value float32) error
	GetStr(tag string, field string) (float32, error)
	SetStr(tag string, field string, value float32) error
}

func New(cfg config.Config) (Cache, error) {
	f, ok := registry[cfg.CurrCache.Type]
	if !ok {
		err := errors.ErrCurrCacheNotFound
		logger.Error(err.Error())
		return nil, err
	}
	return f(cfg)
}
