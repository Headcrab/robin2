package cache

import (
	"context"
	"fmt"
	"net"
	"robin2/internal/config"
	"robin2/internal/logger"
	"strings"
	"sync"
	"time"

	_ "github.com/go-redis/redis"
	"github.com/redis/go-redis/v9"
)

func init() {
	Register("redis", NewRedis)
}

var RedisCacheLock = &sync.Mutex{}

type Redis struct {
	// Cache
	rds    *redis.Client
	config config.Config
	ttl    time.Duration
}

func NewRedis(cfg config.Config) (Cache, error) {
	t := Redis{
		config: cfg,
	}
	err := t.Connect()
	if err != nil {
		logger.Error(err.Error())
		return &t, err
	}
	logger.Trace("NewRedisCache")
	return &t, nil
}

func (c *Redis) Connect() error {
	// cacheName := c.config.CurrCache.Name
	host := c.config.CurrCache.Host
	port := c.config.CurrCache.Port
	password := c.config.CurrCache.Password
	db := c.config.CurrCache.DB
	c.ttl = time.Duration(c.config.CurrCache.TTL) * time.Hour
	logger.Trace("RedisCacheImpl.Connect")
	c.rds = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})
	nips, _ := net.LookupIP(host)
	var ips []string
	for _, ip := range nips {
		ips = append(ips, ip.String())
	}
	logger.Info(fmt.Sprintf("cache connecting to redis on %s:%s ( %s )", host, port, strings.Join(ips, ", ")))
	// ping to check connection
	err := c.rds.Ping(context.Background()).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c Redis) Disconnect() error {
	logger.Trace("RedisCacheImpl.Disconnect")
	return c.rds.Close()
}

func (c Redis) Get(tag string, date time.Time) (float32, error) {
	logger.Trace("RedisCacheImpl.Get")
	c.rds.Expire(context.Background(), tag, c.ttl)
	return c.rds.HGet(context.Background(), tag, date.Format("2006-01-02 15:04:05")).Float32()
}

func (c Redis) Set(tag string, date time.Time, value float32) error {
	logger.Trace("RedisCacheImpl.Set")
	c.rds.Expire(context.Background(), tag, c.ttl)
	c.rds.HSet(context.Background(), tag, date.Format("2006-01-02 15:04:05"), value)
	return nil
}

func (c Redis) GetStr(tag string, field string) (float32, error) {
	logger.Trace("RedisCacheImpl.GetStr")
	c.rds.Expire(context.Background(), tag, c.ttl)
	return c.rds.HGet(context.Background(), tag, field).Float32()

}

func (c Redis) SetStr(tag string, field string, value float32) error {
	logger.Trace("RedisCacheImpl.SetStr")
	c.rds.Expire(context.Background(), tag, c.ttl)
	c.rds.HSet(context.Background(), tag, field, value)
	return nil
}
