package config

import (
	"errors"
	"os"
	"robin2/internal/logger"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	FileName      string
	CurrDB        *Database
	CurrCache     *CacheConfig
	Port          int           `json:"port"`
	Round         int           `json:"round"`
	CurrDBName    string        `json:"curr_db"`
	DB            []Database    `json:"db"`
	CurrCacheName string        `json:"curr_cache"`
	Cache         []CacheConfig `json:"cache"`
	DateFormats   []string      `json:"date_formats"`
}

type Database struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Host             string            `json:"host"`
	Port             string            `json:"port"`
	User             string            `json:"user"`
	Password         string            `json:"password"`
	Database         string            `json:"database"`
	Timeout          int               `json:"timeout"`
	ConnectionString string            `json:"connection_string"`
	Query            map[string]string `json:"query"`
	MaxIdleConns     int               `json:"max_idle_conns,omitempty"`
	MaxOpenConns     int               `json:"max_open_conns,omitempty"`
	ConnMaxIdleTime  int               `json:"conn_max_idle_time,omitempty"`
	ConnMaxLifetime  int               `json:"conn_max_lifetime,omitempty"`
}

type CacheConfig struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	TTL              int    `json:"ttl"`
	Active           string `json:"active"`
	Host             string `json:"host"`
	Port             string `json:"port"`
	Password         string `json:"password"`
	DB               int    `json:"db"`
	MaxOpenConns     int    `json:"max_open_conns"`
	MaxIdleConns     int    `json:"max_idle_conns"`
	ConnMaxLifetime  int    `json:"conn_max_lifetime"`
	ConnectionString string `json:"connection_string"`
}

func (c *Config) Load(fileName string) {
	c.FileName = fileName
	err := c.Reload()
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (c *Config) Reload() error {
	logger.Debug("Initializing config...")
	file, err := os.Open(c.FileName)
	if err != nil {
		logger.Error("logger.Error reading config file")
		return err
	}
	defer file.Close()

	err = cleanenv.ParseJSON(file, c)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	var currDB *Database
	for _, db := range c.DB {
		if db.Name == c.CurrDBName {
			currDB = &db
			break
		}
	}
	if currDB == nil {
		logger.Error("CurrDB not found")
		return errors.New("CurrDB not found")
	} else {
		c.CurrDB = currDB

	}

	var currCache *CacheConfig
	for _, cache := range c.Cache {
		if cache.Name == c.CurrCacheName {
			currCache = &cache
			break
		}
	}
	if currCache == nil {
		logger.Error("CurrCache not found")
		return errors.New("CurrCache not found")
	} else {
		c.CurrCache = currCache
	}
	return nil
}
