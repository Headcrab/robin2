package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	_ "github.com/denisenkom/go-mssqldb"
)

type MsSqlStore interface {
	BaseStore
}

type MsSqlStoreImpl struct {
	MsSqlStore
}

func NewMsSqlStore(cfg config.Config) BaseStore {
	logger.Debug("NewMsSqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	return BaseStore(&MsSqlStoreImpl{
		MsSqlStore: &BaseStoreImpl{
			roundConstant: p,
			config:        cfg,
		},
	})
}

func (s *MsSqlStoreImpl) Connect(name string, cache cache.BaseCache) error {
	logger.Debug("MsSqlStoreImpl.Connect")
	var err error
	base := s.MsSqlStore.(*BaseStoreImpl)
	if base.db != nil {
		err = base.db.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	base.cache = cache
	base.db, err = sql.Open(base.config.CurrDB.Type, base.marshalConnectionString(name))
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	// defer base.db.Close()
	err = base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	base.logConnection(name)
	return nil
}
