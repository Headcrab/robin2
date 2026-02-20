package store

import (
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

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
	return s.Base.connectWithReuse(name, cache)
}
