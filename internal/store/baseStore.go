package store

import (
	"database/sql"
	"math"
	"robin2/internal/errors"
	"strconv"
	"strings"
	"time"

	"robin2/internal/cache"
	"robin2/pkg/config"
	"robin2/pkg/logger"
	// _ "github.com/denisenkom/go-mssqldb"
	// _ "github.com/go-sql-driver/mysql"
)

/*------------------------------------------------------------------*/

type BaseStore interface {
	Connect(cache cache.BaseCache) error
	GetTagDate(tag string, date time.Time) (float32, error)
	GetTagCount(tag string, from time.Time, to time.Time, strCount int) (map[string]map[time.Time]float32, error)
	GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error)
	GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) string
	GetTagList(like string) (map[string][]string, error)
	GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error)

	RoundAndFormat(val float32) string
	Round(val float32) float32
}

type BaseStoreImpl struct {
	BaseStore
	db     *sql.DB
	config config.Config
	cache  cache.BaseCache
}

/*------------------------------------------------------------------*/

func (s *BaseStoreImpl) marshalConnectionString() string {
	connStr := s.config.GetString("db." + s.config.GetString("app.db.name") + ".connection_string")
	for k, v := range s.config.GetStringMapString("db." + s.config.GetString("app.db.name")) {
		connStr = strings.ReplaceAll(connStr, "{"+k+"}", v)
	}
	return connStr
}

func (s *BaseStoreImpl) RoundAndFormat(val float32) string {
	f := float64(val)
	round := s.config.GetInt("app.round")
	p := math.Pow(10, float64(round))
	rounded := math.Round(f*p) / p
	ret := strconv.FormatFloat(rounded, 'f', round, 64)
	ret = strings.Replace(ret, ".", ",", -1)
	return ret
}

func (s *BaseStoreImpl) Round(val float32) float32 {
	round := float64(s.config.GetInt("app.round"))
	return float32(math.Round(float64(val)*math.Pow(10, round)) / math.Pow(10, round))
}

func (s *BaseStoreImpl) replaceTemplate(repMap map[string]string, query string) string {
	for k, v := range repMap {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

func (s *BaseStoreImpl) GetTagDate(tag string, date time.Time) (float32, error) {
	// logger.Log(logger.Trace, "GetTagDate "+tag+" : "+date.Format("2006-01-02 15:04:05"))
	if date.IsZero() {
		return 0, errors.ErrInvalidDate
	}
	if s.db == nil {
		return 0, errors.ErrDbConnectionFailed
	}
	if s.cache != nil {
		if val, err := s.cache.Get(tag, date); err == nil {
			return val, nil
		}
	}
	query := s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_tag_date")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{date}": date.Format("2006-01-02 15:04:05")}, query)

	var val float32
	err := s.db.QueryRow(query).Scan(&val)
	// defer s.db.Close()
	if err != nil {
		return -1, err
	}
	if s.cache != nil && val != -1 {
		s.cache.Set(tag, date, val)
	}
	// logger.Log(logger.Trace, "GetTagDate "+tag+" : "+date.Format("2006-01-02 15:04:05")+" from db")
	return val, nil
}

func (s *BaseStoreImpl) GetTagCount(tag string, from time.Time, to time.Time, count int) (map[string]map[time.Time]float32, error) {
	logger.Log(logger.Debug, "GetTagCount "+tag+" : "+from.Format("2006-01-02 15:04:05")+" - "+to.Format("2006-01-02 15:04:05")+" ("+strconv.Itoa(count)+")")
	if count == 0 {
		return nil, errors.ErrCountIsEmpty
	}
	if count < 1 {
		return nil, errors.ErrCountIsLessThanOne
	}

	tmDiff := (time.Until(to).Seconds() - time.Until(from).Seconds()) / float64(count)
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
			// if s.cache != nil {
			// 	if val, err := s.cache.Get(t, dateFrom); err == nil {
			// 		// logger.Log(logger.Trace, "GetTagDate "+t+" : "+dateFrom.Format("2006-01-02 15:04:05")+" from cache")
			// 		resDt[dateFrom] = s.Round(val)
			// 		continue
			// 	}
			// }
			val, err := s.GetTagDate(t, dateFrom)
			if err != nil {
				val = -1
			}
			resDt[dateFrom] = s.Round(val)
			// if s.cache != nil && val != -1 {
			// 	s.cache.Set(t, dateFrom, val)
			// }
		}
		res[t] = resDt
	}
	return res, nil
}

func (s *BaseStoreImpl) GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error) {
	logger.Log(logger.Debug, "GetTagFromTo "+tag+" : "+from.Format("2006-01-02 15:04:05")+" - "+to.Format("2006-01-02 15:04:05"))
	count := int(time.Until(to).Seconds() - time.Until(from).Seconds())
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(time.Second * time.Duration((i))))
			// if s.cache != nil {
			// 	if val, err := s.cache.Get(tag, dateFrom); err == nil {
			// 		// logger.Log(logger.Trace, "GetTagDate "+tag+" : "+dateFrom.Format("2006-01-02 15:04:05")+" from cache")
			// 		resDt[dateFrom] = s.Round(val)
			// 		continue
			// 	}
			// }
			val, err := s.GetTagDate(t, dateFrom)
			if err != nil {
				val = -1
			}
			resDt[dateFrom] = s.Round(val)
			// if s.cache != nil && val != -1 {
			// 	s.cache.Set(tag, dateFrom, val)
			// }
		}
		res[t] = resDt
	}
	return res, nil
}

func (s *BaseStoreImpl) GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) string {
	logger.Log(logger.Debug, "GetTagFromTo "+tag+" : "+from.Format("2006-01-02 15:04:05")+" - "+to.Format("2006-01-02 15:04:05")+" ("+group+")")
	var query string
	group = strings.ToLower(group)
	if group == "avg" || group == "sum" || group == "min" || group == "max" {
		query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_tag_from_to_group")
	}

	if group == "dif" {
		query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_tag_from_to_dif")
	}

	if group == "count" {
		query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_tag_from_to_count")
	}
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")

	// todo: cache
	// if s.cache != nil {
	// 	if val, err := (*s.cache).Get(tag, date); err == nil {
	// 		return val, nil
	// 	}
	// }

	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr, "{group}": group}, query)
	if query == "" {
		return "#Error: group error"
	}
	var value float32
	err := s.db.QueryRow(query).Scan(&value)
	if err != nil {
		return "#Error: tag or date not found (" + err.Error() + ")"
	}

	// todo: cache
	// if s.cache != nil {
	// 	(*s.cache).Set(tag, date, value)
	// }

	// sleep 5 seconds
	// time.Sleep(2 * time.Second)
	return s.RoundAndFormat(value)
}

func (s *BaseStoreImpl) GetTagList(like string) (map[string][]string, error) {
	if like == "" {
		like = "%"
	}
	like = s.replaceTemplate(map[string]string{"*": "%", "?": "_", " ": "%"}, like)
	query := s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_tag_list")
	// replace {tag} with like
	query = strings.Replace(query, "{tag}", like, -1)
	tags := make(map[string][]string)
	cur, err := s.db.Query(query)
	if err != nil {
		logger.Log(logger.Debug, err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var tag string
			err := cur.Scan(&tag)
			if err != nil {
				logger.Log(logger.Debug, err.Error())
				return nil, err
			} else {
				tags["tags"] = append(tags["tags"], tag)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Log(logger.Debug, err.Error())
		}
	}()
	return tags, nil
}

func (s *BaseStoreImpl) GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Log(logger.Debug, "GetDownDate "+tag+" : "+from.Format("2006-01-02 15:04:05")+" - "+to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_down_dates")
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.ErrQueryError
	}
	var dates []time.Time
	cur, err := s.db.Query(query)
	if err != nil {
		logger.Log(logger.Debug, err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var date time.Time
			err := cur.Scan(&date)
			if err != nil {
				logger.Log(logger.Debug, err.Error())
				return nil, err
			} else {
				dates = append(dates, date)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Log(logger.Debug, err.Error())
		}
	}()
	return dates, nil
}

func (s *BaseStoreImpl) GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Log(logger.Debug, "GetUpDate "+tag+" : "+from.Format("2006-01-02 15:04:05")+" - "+to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_up_dates")
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.ErrQueryError
	}
	var dates []time.Time
	cur, err := s.db.Query(query)
	if err != nil {
		logger.Log(logger.Debug, err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var date time.Time
			err := cur.Scan(&date)
			if err != nil {
				logger.Log(logger.Debug, err.Error())
				return nil, err
			} else {
				dates = append(dates, date)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Log(logger.Debug, err.Error())
		}
	}()
	return dates, nil
}
