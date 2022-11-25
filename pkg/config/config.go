package config

import (
	"robin2/pkg/logger"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
}

var config *Config
var lock = &sync.Mutex{}

func GetConfig() *Config {
	if config == nil {
		config = &Config{}
		config.Init()
	}
	return config
}

func (c *Config) Init() {
	logger.Log(logger.Debug, "Initializing config...")
	lock.Lock()
	defer lock.Unlock()
	viper.SetConfigName("robin2.cfg")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("../../bin/configs")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Log(logger.Error, "error reading config file")
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
	logger.Log(logger.Debug, "Reloading config...")
	lock.Lock()
	defer lock.Unlock()
	err = viper.ReadInConfig()
	if err != nil {
		logger.Log(logger.Error, "logger.Error reading config file")
	}
	return err
}
