package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type Clickhouse struct {
	Base
}

func NewClickhouse(cfg config.Config) *Clickhouse {
	logger.Debug("NewClickHouseStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := Clickhouse{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t
}

func (s *Clickhouse) Connect(name string, cache cache.Cache) error {
	logger.Debug("ClickHouseStoreImpl.Connect")

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	s.cache = cache

	var err error
	s.db, err = sql.Open(s.config.CurrDB.Type, s.marshalConnectionString(name))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if err = s.db.Ping(); err != nil {
		logger.Error(err.Error())
		return err
	}

	s.logConnection(name)

	return nil
}
