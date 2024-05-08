package store

import (
	"time"

	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/data"
	"robin2/internal/errors"
)

var registry map[string]func(config.Config) (Store, error)

func Register(name string, f func(config.Config) (Store, error)) {
	if registry == nil {
		registry = make(map[string]func(config.Config) (Store, error))
	}
	registry[name] = f
}

func New(cfg config.Config) (Store, error) {
	f, ok := registry[cfg.CurrDB.Type]
	if !ok {
		err := errors.ErrCurrDBNotFound
		// logger.Error(err.Error())
		return nil, err
	}
	return f(cfg)
}

type Store interface {
	Connect(name string, cache cache.Cache) error
	GetTagDate(tag string, date time.Time) (*data.Tag, error)
	// GetTagsDate(tags []string, date time.Time) (, error)
	GetTagCount(tag string, from time.Time, to time.Time, strCount int) (map[string]map[time.Time]float32, error)
	GetTagCountGroup(tag string, from time.Time, to time.Time, strCount int, group string) (data.Tags, error)
	GetTagFromTo(tag string, from time.Time, to time.Time) (data.Tags, error)
	GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error)
	GetTagList(like string) (*data.Output, error)
	GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetStatus() (string, time.Duration, error)

	TemplateList(like string) (map[string]string, error)
	TemplateExec(name string, params map[string]string) (*data.Output, error)

	TemplateAdd(name string, body string) error
	TemplateSet(name string, body string) error
	TemplateGet(name string) (string, error)
	TemplateDel(name string) error

	ExecQuery(query string) (*data.Output, error)
}
