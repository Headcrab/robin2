package store

import "robin2/pkg/logger"

type Factory interface {
	NewStore(string) BaseStore
}

func NewFactory() Factory {
	logger.Log(logger.Debug, "NewStoreFactory")
	return &FactoryImpl{}
}

type FactoryImpl struct {
	Factory
}

func (f *FactoryImpl) NewStore(dbName string) BaseStore {
	switch dbName {
	case "mysql":
		logger.Log(logger.Debug, "NewStoreFactory.NewStore.mysql")
		return NewMySqlStore()
	case "mssql":
		logger.Log(logger.Debug, "NewStoreFactory.NewStore.mssql")
		return NewMsSqlStore()
	case "clickhouse":
		logger.Log(logger.Debug, "NewStoreFactory.NewStore.clickhouse")
		return NewClickHouseStore()
	default:
		logger.Log(logger.Error, "NewStoreFactory.NewStore.default: "+dbName)
		return nil
	}
}
