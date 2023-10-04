package store

import (
	"database/sql"
	"math"
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

func NewClickHouseStore() BaseStore {
	logger.Debug("NewClickHouseStore")
	conf := config.GetConfig()
	round := conf.GetInt("app.round")
	p := math.Pow(10, float64(round))
	return BaseStore(&ClickHouseStoreImpl{
		ClickHouseStore: &BaseStoreImpl{
			roundConstant: p,
			config:        config.GetConfig(),
		},
	})
}

func (s *ClickHouseStoreImpl) Connect(cache cache.BaseCache) error {
	logger.Debug("ClickHouseStoreImpl.Connect")

	base := s.ClickHouseStore.(*BaseStoreImpl)
	if base.db != nil {
		if err := base.db.Close(); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	base.cache = cache

	dbType := base.config.GetString("app.db.type")

	db, err := sql.Open(dbType, base.marshalConnectionString())
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if err = db.Ping(); err != nil {
		logger.Error(err.Error())
		return err
	}

	base.db = db

	base.logConnection()

	return nil
}
