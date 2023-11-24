package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlStore interface {
	BaseStore
}

type MySqlStoreImpl struct {
	MySqlStore
}

func NewMySqlStore(cfg config.Config) BaseStore {
	logger.Debug("NewMySqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	return BaseStore(&MySqlStoreImpl{
		MySqlStore: &BaseStoreImpl{
			roundConstant: p,
			config:        cfg,
		},
	})
}

func (s *MySqlStoreImpl) Connect(name string, cache cache.BaseCache) error {
	logger.Debug("MySqlStoreImpl.Connect")
	var err error
	base := s.MySqlStore.(*BaseStoreImpl)
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
	base.db.SetMaxIdleConns(base.config.CurrDB.MaxIdleConns)
	base.db.SetMaxOpenConns(base.config.CurrDB.MaxOpenConns)
	base.db.SetConnMaxIdleTime(time.Duration(base.config.CurrDB.ConnMaxIdleTime) * time.Second)
	base.db.SetConnMaxLifetime(time.Duration(base.config.CurrDB.ConnMaxLifetime) * time.Second)
	// todo: CHECK! setup strings
	// for _, v := range base.config.CurrDB.SetUpStrings {
	// 	_, err = base.db.Exec(v)
	// 	if err != nil {
	// 		logger.Error(err.Error())
	// 		return err
	// 	}
	// }
	// defer base.db.Close()
	err = base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	base.logConnection(name)
	return nil
}
