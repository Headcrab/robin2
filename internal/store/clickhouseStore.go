package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseStore interface {
	BaseStore
}

type ClickHouseStoreImpl struct {
	ClickHouseStore
}

func NewClickHouseStore(cfg config.Config) BaseStore {
	logger.Debug("NewClickHouseStore")
	round := cfg.GetInt("app.round")
	p := math.Pow(10, float64(round))
	return BaseStore(&ClickHouseStoreImpl{
		ClickHouseStore: &BaseStoreImpl{
			roundConstant: p,
			config:        cfg,
		},
	})
}

func (s *ClickHouseStoreImpl) Connect(name string, cache cache.BaseCache) error {
	logger.Debug("ClickHouseStoreImpl.Connect")

	base := s.ClickHouseStore.(*BaseStoreImpl)
	if base.db != nil {
		if err := base.db.Close(); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	base.cache = cache

	dbType := base.config.GetString("app.db." + name + ".type")

	db, err := sql.Open(dbType, base.marshalConnectionString(name))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if err = db.Ping(); err != nil {
		logger.Error(err.Error())
		return err
	}

	base.db = db

	base.logConnection(name)

	return nil
}
