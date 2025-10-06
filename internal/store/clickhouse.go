package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func init() {
	Register("clickhouse", NewClickhouse)
}

type Clickhouse struct {
	Base
}

func NewClickhouse(cfg config.Config) (Store, error) {
	logger.Debug("NewClickHouseStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := Clickhouse{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t, nil
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
	s.db, err = sql.Open(s.config.CurrDB.Type, s.GenerateConnectionString(name))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	// Устанавливаем лимит соединений
	maxConnLimit := s.getMaxConnLimit()
	s.db.SetMaxOpenConns(maxConnLimit)

	// Устанавливаем другие параметры если они заданы
	if s.config.CurrDB.MaxIdleConns > 0 {
		s.db.SetMaxIdleConns(s.config.CurrDB.MaxIdleConns)
	}
	if s.config.CurrDB.ConnMaxIdleTime > 0 {
		s.db.SetConnMaxIdleTime(time.Duration(s.config.CurrDB.ConnMaxIdleTime) * time.Second)
	}
	if s.config.CurrDB.ConnMaxLifetime > 0 {
		s.db.SetConnMaxLifetime(time.Duration(s.config.CurrDB.ConnMaxLifetime) * time.Second)
	}

	if err = s.db.Ping(); err != nil {
		logger.Error(err.Error())
		return err
	}

	s.logConnection(name)

	return nil
}
