package store

import (
	"database/sql"
	"robin2/internal/cache"
	"robin2/pkg/config"
	"robin2/pkg/logger"
)

type MySqlStore interface {
	BaseStore
}

type MySqlStoreImpl struct {
	MySqlStore
}

func NewMySqlStore() *BaseStore {
	logger.Log(logger.Debug, "NewMySqlStore")
	t := BaseStore(&MySqlStoreImpl{
		MySqlStore: &BaseStoreImpl{
			config: *config.GetConfig(),
		},
	})
	return &t
}

func (s *MySqlStoreImpl) Connect(cache *cache.BaseCache) error {
	logger.Log(logger.Debug, "MySqlStoreImpl.Connect")
	var err error
	base := s.MySqlStore.(*BaseStoreImpl)
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
