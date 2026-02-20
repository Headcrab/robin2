package store

import (
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	Register("mysql", NewMySql)
}

type MySql struct {
	Base
}

func NewMySql(cfg config.Config) (Store, error) {
	logger.Debug("NewMySqlStore")
	round := cfg.Round
	p := math.Pow(10, float64(round))
	t := MySql{
		Base: Base{
			roundConstant: p,
			config:        cfg,
		},
	}
	return &t, nil
}

func (s *MySql) Connect(name string, cache cache.Cache) error {
	logger.Debug("MySqlStoreImpl.Connect")
	return s.Base.connectWithReuse(name, cache)
}
