package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func init() {
	Register("mssql", NewMsSql)
}

type MsSql struct {
	Base
}

func NewMsSql(cfg config.Config) (Store, error) {
	logger.Debug("NewMsSqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := MsSql{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t, nil
}

func (s *MsSql) Connect(name string, cache cache.Cache) error {
	logger.Debug("MsSqlStoreImpl.Connect")
	var err error
	if s.Base.db != nil {
		err = s.Base.db.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	s.cache = cache
	s.Base.db, err = sql.Open(s.Base.config.CurrDB.Type, s.Base.GenerateConnectionString(name))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	// Устанавливаем лимит соединений с приоритетом на MaxConnLimit
	maxConnLimit := s.getMaxConnLimit()

	// Если MaxOpenConns задан в конфиге, используем минимальное из двух значений
	if s.Base.config.CurrDB.MaxOpenConns > 0 {
		if s.Base.config.CurrDB.MaxOpenConns < maxConnLimit {
			maxConnLimit = s.Base.config.CurrDB.MaxOpenConns
		}
	}
	s.Base.db.SetMaxOpenConns(maxConnLimit)

	// Устанавливаем остальные параметры
	if s.Base.config.CurrDB.MaxIdleConns > 0 {
		s.Base.db.SetMaxIdleConns(s.Base.config.CurrDB.MaxIdleConns)
	}
	if s.Base.config.CurrDB.ConnMaxIdleTime > 0 {
		s.Base.db.SetConnMaxIdleTime(time.Duration(s.Base.config.CurrDB.ConnMaxIdleTime) * time.Second)
	}
	if s.Base.config.CurrDB.ConnMaxLifetime > 0 {
		s.Base.db.SetConnMaxLifetime(time.Duration(s.Base.config.CurrDB.ConnMaxLifetime) * time.Second)
	}

	// defer base.db.Close()
	err = s.Base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	s.Base.logConnection(name)
	return nil
}
