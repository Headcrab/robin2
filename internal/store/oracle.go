package store

import (
	"database/sql"
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"
	"time"

	_ "github.com/sijms/go-ora/v2"
)

func init() {
	Register("oracle", NewOracle)
}

type Oracle struct {
	Base
}

func NewOracle(cfg config.Config) (Store, error) {
	logger.Debug("NewOracleStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := Oracle{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t, nil
}

func (s *Oracle) Connect(name string, cache cache.Cache) error {
	logger.Debug("OracleStoreImpl.Connect")
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
	s.Base.db.SetMaxIdleConns(s.Base.config.CurrDB.MaxIdleConns)
	s.Base.db.SetMaxOpenConns(s.Base.config.CurrDB.MaxOpenConns)
	s.Base.db.SetConnMaxIdleTime(time.Duration(s.Base.config.CurrDB.ConnMaxIdleTime) * time.Second)
	s.Base.db.SetConnMaxLifetime(time.Duration(s.Base.config.CurrDB.ConnMaxLifetime) * time.Second)
	// todo: CHECK! setup strings
	// for _, v := range base.config.CurrDB.SetUpStrings {
	// 	_, err = base.db.Exec(v)
	// 	if err != nil {
	// 		logger.Error(err.Error())
	// 		return err
	// 	}
	// }
	// defer base.db.Close()
	err = s.Base.db.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	s.Base.logConnection(name)
	return nil
}
