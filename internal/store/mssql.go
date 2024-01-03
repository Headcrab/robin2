package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	_ "github.com/denisenkom/go-mssqldb"
)

type MsSql struct {
	Base
}

func NewMsSql(cfg config.Config) *MsSql {
	logger.Debug("NewMsSqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := MsSql{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t
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
	// defer base.db.Close()
	err = s.Base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	s.Base.logConnection(name)
	return nil
}
