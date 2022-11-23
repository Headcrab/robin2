package cache

import (
	"time"
)

type BaseCache interface {
	Connect() error
	Disconnect() error
	Get(tag string, date time.Time) (float32, error)
	Set(tag string, date time.Time, value float32) error
	// Remove(key string) error
	// RemoveHash(key hash) error
	// RemoveAll() error
}

type BaseCacheImpl struct {
	BaseCache
}

func Connect() error {
	return nil
}

func Disconnect() error {
	return nil
}
