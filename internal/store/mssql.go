package store

import (
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

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
	return s.Base.connectWithReuse(name, cache)
}
