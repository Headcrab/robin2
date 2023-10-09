package store

import (
	"robin2/internal/config"
	"robin2/internal/logger"
)

type Factory interface {
	NewStore(string, config.Config) BaseStore
}

func NewFactory() Factory {
	logger.Debug("NewStoreFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewStore(dbName string, cfg config.Config) BaseStore {
	switch dbName {
	case "mysql":
		logger.Debug("NewStoreFactory.NewStore.mysql")
		return NewMySqlStore(cfg)
	case "mssql":
		logger.Debug("NewStoreFactory.NewStore.mssql")
		return NewMsSqlStore(cfg)
	case "clickhouse":
		logger.Debug("NewStoreFactory.NewStore.clickhouse")
		return NewClickHouseStore(cfg)
	default:
		logger.Error("NewStoreFactory.NewStore.default: " + dbName)
		return nil
	}
}
