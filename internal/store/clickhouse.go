package store

import (
	"math"
	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

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
	return s.Base.connectWithReuse(name, cache)
}
