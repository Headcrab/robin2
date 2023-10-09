package config

import (
	"path/filepath"
	"robin2/internal/logger"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
}

var config *Config
var lock = &sync.Mutex{}

func GetConfig(path string) Config {
	if config == nil {
		config = &Config{}
		config.Load(filepath.Join(path, "config"))
	}
	return *config
}

func (c *Config) Load(path string) {
	logger.Debug("Initializing config...")
	lock.Lock()
	defer lock.Unlock()
	viper.SetConfigName("Robin.json")
	viper.AddConfigPath(path)
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("error reading config file")
	}
}

func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

func (c *Config) GetInt(key string) int {
	return viper.GetInt(key)
}

func (c *Config) GetStringMapString(key string) map[string]string {
	return viper.GetStringMapString(key)
}

func (c *Config) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (c *Config) GetAllSettings() map[string]interface{} {
	return viper.AllSettings()
}

func (c *Config) Reload() (err error) {
	logger.Debug("Reloading config...")
	lock.Lock()
	defer lock.Unlock()
	err = viper.ReadInConfig()
	if err != nil {
		logger.Error("logger.Error reading config file")
	}
	return err
}
