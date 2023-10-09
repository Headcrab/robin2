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

func NewMySqlStore() BaseStore {
	logger.Debug("NewMySqlStore")
	conf := config.GetConfig()
	round := conf.GetInt("app.round")
	p := math.Pow(10, float64(round))
	return BaseStore(&MySqlStoreImpl{
		MySqlStore: &BaseStoreImpl{
			roundConstant: p,
			config:        config.GetConfig(),
		},
	})
}

func (s *MySqlStoreImpl) Connect(cache cache.BaseCache) error {
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
	base.db, err = sql.Open(base.config.GetString("app.db.type"), base.marshalConnectionString())
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	base.db.SetMaxIdleConns(base.config.GetInt("app.db." + base.config.GetString("app.db.current") + ".max_idle_conns"))
	base.db.SetMaxOpenConns(base.config.GetInt("app.db." + base.config.GetString("app.db.current") + ".max_open_conns"))
	base.db.SetConnMaxIdleTime(time.Duration(base.config.GetInt("app.db."+base.config.GetString("app.db.current")+".conn_max_idle_time")) * time.Second)
	base.db.SetConnMaxLifetime(time.Duration(base.config.GetInt("app.db."+base.config.GetString("app.db.current")+".conn_max_lifetime")) * time.Second)
	// setup strings
	for _, v := range base.config.GetStringSlice("app.db." + base.config.GetString("app.db.current") + ".setup_strings") {
		_, err = base.db.Exec(v)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	// defer base.db.Close()
	err = base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	base.logConnection()
	return nil
}
