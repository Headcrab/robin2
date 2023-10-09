package store

import "robin2/internal/logger"

type Factory interface {
	NewStore(string) BaseStore
}

func NewFactory() Factory {
	logger.Debug("NewStoreFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewStore(dbName string) BaseStore {
	switch dbName {
	case "mysql":
		logger.Debug("NewStoreFactory.NewStore.mysql")
		return NewMySqlStore()
	case "mssql":
		logger.Debug("NewStoreFactory.NewStore.mssql")
		return NewMsSqlStore()
	case "clickhouse":
		logger.Debug("NewStoreFactory.NewStore.clickhouse")
		return NewClickHouseStore()
	default:
		logger.Error("NewStoreFactory.NewStore.default: " + dbName)
		return nil
	}
}
