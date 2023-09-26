package cache

import (
	"context"
	"fmt"
	"robin2/pkg/config"
	"robin2/pkg/logger"
	"sync"
	"time"

	_ "github.com/go-redis/redis"
	"github.com/redis/go-redis/v9"
)

var RedisCacheLock = &sync.Mutex{}

type RedisCacheImpl struct {
	BaseCache
	rds    *redis.Client
	config config.Config
}

func NewRedisCache() (BaseCache, error) {
	t := BaseCache(&RedisCacheImpl{
		config: config.GetConfig(),
	})
	err := t.Connect()
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	logger.Trace("NewRedisCache")
	return t, nil
}

func (c *RedisCacheImpl) Connect() error {
	cacheName := c.config.GetString("app.cache.name")
	host := c.config.GetString("db." + cacheName + ".host")
	port := c.config.GetString("db." + cacheName + ".port")
	password := c.config.GetString("db." + cacheName + ".password")
	db := c.config.GetInt("db." + cacheName + ".db")

	logger.Trace("RedisCacheImpl.Connect")
	c.rds = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})
	logger.Info(fmt.Sprintf("cache connected to redis on %s:%s", host, port))
	// ping to check connection
	err := c.rds.Ping(context.Background()).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisCacheImpl) Get(tag string, date time.Time) (float32, error) {
	logger.Trace("RedisCacheImpl.Get")
	c.rds.Expire(context.Background(), tag, time.Duration(c.config.GetInt("app.cache.ttl"))*time.Hour)
	return c.rds.HGet(context.Background(), tag, date.Format("02.01.2006 15:04:05")).Float32()
}

func (c *RedisCacheImpl) Set(tag string, date time.Time, value float32) error {
	logger.Trace("RedisCacheImpl.Set")
	// RedisCacheLock.Lock()
	// defer RedisCacheLock.Unlock()

	// устанавливаем TTL
	c.rds.Expire(context.Background(), tag, time.Duration(c.config.GetInt("app.cache.ttl"))*time.Hour)
	c.rds.HSet(context.Background(), tag, date.Format("02.01.2006 15:04:05"), value)
	return nil
}
