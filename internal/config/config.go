package config

import (
	// "errors"
	"fmt"
	"os"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"strings"

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
	MaxConnLimit     int               `json:"max_conn_limit,omitempty"`
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

func New() *Config {
	return &Config{}
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
	c.expandEnvRefs()

	var currDB *Database
	for i := range c.DB {
		if c.DB[i].Name == c.CurrDBName {
			currDB = &c.DB[i]
			break
		}
	}
	if currDB == nil {
		logger.Error(fmt.Sprintf("CurrDB '%s' not found", c.CurrDBName))
		return errors.ErrCurrDBNotFound
	} else {
		c.CurrDB = currDB

	}

	var currCache *CacheConfig
	for i := range c.Cache {
		if c.Cache[i].Name == c.CurrCacheName {
			currCache = &c.Cache[i]
			break
		}
	}
	if currCache == nil {
		logger.Error(fmt.Sprintf("CurrCache '%s' not found", c.CurrCacheName))
		return errors.ErrCurrCacheNotFound
	} else {
		c.CurrCache = currCache
	}
	return nil
}

func (c *Config) expandEnvRefs() {
	for i := range c.DB {
		c.DB[i].Host = expandEnvRef(c.DB[i].Host)
		c.DB[i].Port = expandEnvRef(c.DB[i].Port)
		c.DB[i].User = expandEnvRef(c.DB[i].User)
		c.DB[i].Password = expandEnvRef(c.DB[i].Password)
		c.DB[i].Database = expandEnvRef(c.DB[i].Database)
		c.DB[i].ConnectionString = expandEnvRef(c.DB[i].ConnectionString)
	}

	for i := range c.Cache {
		c.Cache[i].Host = expandEnvRef(c.Cache[i].Host)
		c.Cache[i].Port = expandEnvRef(c.Cache[i].Port)
		c.Cache[i].Password = expandEnvRef(c.Cache[i].Password)
		c.Cache[i].ConnectionString = expandEnvRef(c.Cache[i].ConnectionString)
	}
}

func expandEnvRef(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "${") && strings.HasSuffix(trimmed, "}") {
		key := strings.TrimSuffix(strings.TrimPrefix(trimmed, "${"), "}")
		if envValue, ok := os.LookupEnv(key); ok {
			return envValue
		}
	}
	return value
}

func isEnvRef(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.HasPrefix(trimmed, "${") && strings.HasSuffix(trimmed, "}")
}

func (c *Config) ValidateCurrent() error {
	if c.CurrDB == nil {
		return errors.ErrCurrDBNotFound
	}
	if isEnvRef(c.CurrDB.Host) || isEnvRef(c.CurrDB.User) || isEnvRef(c.CurrDB.Password) {
		return fmt.Errorf("current database %q requires environment secrets", c.CurrDB.Name)
	}
	if c.CurrCache != nil && isEnvRef(c.CurrCache.Host) {
		return fmt.Errorf("current cache %q requires environment secrets", c.CurrCache.Name)
	}
	return nil
}
