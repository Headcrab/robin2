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

type MySql struct {
	Base
}

func NewMySql(cfg config.Config) *MySql {
	logger.Debug("NewMySqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := MySql{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t
}

func (s *MySql) Connect(name string, cache cache.Cache) error {
	logger.Debug("MySqlStoreImpl.Connect")
	var err error
	if s.Base.db != nil {
		err = s.Base.db.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	s.cache = cache
	s.Base.db, err = sql.Open(s.Base.config.CurrDB.Type, s.Base.marshalConnectionString(name))
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
