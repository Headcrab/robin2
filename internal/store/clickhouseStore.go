package store

import (
	"database/sql"
	"robin2/internal/cache"
	"robin2/pkg/config"
	"robin2/pkg/logger"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseStore interface {
	BaseStore
}

type ClickHouseStoreImpl struct {
	ClickHouseStore
}

func NewClickHouseStore() *BaseStore {
	logger.Log(logger.Debug, "NewClickHouseStore")
	t := BaseStore(&ClickHouseStoreImpl{
		ClickHouseStore: &BaseStoreImpl{
			config: *config.GetConfig(),
		},
	})
	return &t
}

func (s *ClickHouseStoreImpl) Connect(cache cache.BaseCache) error {
	logger.Log(logger.Debug, "ClickHouseStoreImpl.Connect")
	var err error
	base := s.ClickHouseStore.(*BaseStoreImpl)
	if base.db != nil {
		err = base.db.Close()
		if err != nil {
			logger.Log(logger.Error, err.Error())
		}
	}
	base.cache = cache
	base.db, err = sql.Open(base.config.GetString("app.db.type"), base.marshalConnectionString())
	if err != nil {
		logger.Log(logger.Error, err.Error())
		return err
	}
	// defer base.db.Close()
	err = base.db.Ping()
	if err != nil {
		logger.Log(logger.Error, err.Error())
		return err
	}
	logger.Log(logger.Info, "connected to "+base.config.GetString("app.db.type")+" database on "+
		base.config.GetString("db."+base.config.GetString("app.db.name")+".host")+":"+
		base.config.GetString("db."+base.config.GetString("app.db.name")+".port"))
	return nil
}
