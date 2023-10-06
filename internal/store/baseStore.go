package store

import (
	"database/sql"
	"fmt"
	"math"
	"net"
	"robin2/internal/errors"
	"strconv"
	"strings"
	"time"

	"robin2/internal/cache"
	"robin2/pkg/config"
	"robin2/pkg/logger"

	"github.com/google/uuid"
	// _ "github.com/denisenkom/go-mssqldb"
	// _ "github.com/go-sql-driver/mysql"
)

/*------------------------------------------------------------------*/

type BaseStore interface {
	Connect(cache cache.BaseCache) error
	GetTagDate(tag string, date time.Time) (float32, error)
	GetTagCount(tag string, from time.Time, to time.Time, strCount int) (map[string]map[time.Time]float32, error)
	GetTagCountGroup(tag string, from time.Time, to time.Time, strCount int, group string) (map[string]map[time.Time]float32, error)
	GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error)
	GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error)
	GetTagList(like string) (map[string][]string, error)
	GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetStatus() (string, string, error)

	TemplateList(like string) (map[string]string, error)
	TemplateExec(name string, params map[string]string) (string, error)
	TemplateAdd(name string, body string) error
	TemplateSet(name string, body string) error
	TemplateGet(name string) (string, error)
	TemplateDel(name string) error

	SetRound(round int)
	Format(val float32) string
	Round(val float32) float32
}

type BaseStoreImpl struct {
	BaseStore
	db            *sql.DB
	config        config.Config
	cache         cache.BaseCache
	roundConstant float64
	round         int
}

/*------------------------------------------------------------------*/

func (s *BaseStoreImpl) marshalConnectionString() string {
	connStr := s.config.GetString("db." + s.config.GetString("app.db.name") + ".connection_string")
	for k, v := range s.config.GetStringMapString("db." + s.config.GetString("app.db.name")) {
		connStr = strings.ReplaceAll(connStr, "{"+k+"}", v)
	}
	s.round = s.config.GetInt("app.round")
	return connStr
}

func (s *BaseStoreImpl) logConnection() {
	dbName := s.config.GetString("app.db.name")
	dbType := s.config.GetString("app.db.type")
	host := s.config.GetString("db." + dbName + ".host")
	port := s.config.GetString("db." + dbName + ".port")
	nips, _ := net.LookupIP(host)
	var ips []string
	for _, ip := range nips {
		ips = append(ips, ip.String())
	}
	logger.Info(fmt.Sprintf("connected to %s database on %s:%s ( %s )", dbType, host, port, strings.Join(ips, ", ")))
}

func (s *BaseStoreImpl) SetRound(round int) {
	s.round = round
}

func (s *BaseStoreImpl) Format(val float32) string {
	// f := float64(val)
	// p := math.Pow(10, float64(s.round))
	// rounded := math.Round(f*p) / p
	// ret := strconv.FormatFloat(rounded, 'f', s.round, 64)
	ret := strconv.FormatFloat(float64(val), 'f', s.round, 64)
	ret = strings.Replace(ret, ".", ",", -1)
	return ret
}

func (s *BaseStoreImpl) Round(val float32) float32 {
	// round := float64(s.config.GetInt("app.round"))
	return float32(math.Round(float64(val)*math.Pow(10, float64(s.round))) / math.Pow(10, float64(s.round)))
}

func (s *BaseStoreImpl) replaceTemplate(repMap map[string]string, query string) string {
	for k, v := range repMap {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

func (s *BaseStoreImpl) GetStatus() (string, string, error) {
	var version, uptime string
	err := s.db.QueryRow(s.config.GetString("db."+s.config.GetString("app.db.name")+".query.status")).Scan(&version, &uptime)
	if err != nil {
		return "", "", err
	}
	return version, uptime, nil
}

func (s *BaseStoreImpl) GetTagDate(tag string, date time.Time) (float32, error) {
	if date.IsZero() {
		return 0, errors.InvalidDate
	}
	if s.db == nil {
		return 0, errors.DbConnectionFailed
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
	if err != nil {
		return -1, err
	}
	if s.cache != nil && val != -1 {
		s.cache.Set(tag, date, val)
	}
	return val, nil
}

func (s *BaseStoreImpl) GetTagCount(tag string, from time.Time, to time.Time, count int) (map[string]map[time.Time]float32, error) {
	logger.Debug(fmt.Sprintf("GetTagCount %s : %s - %s (%d)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), count))
	if count == 0 {
		return nil, errors.CountIsEmpty
	}
	if count < 1 {
		return nil, errors.CountIsLessThanOne
	}

	tmDiff := to.Sub(from).Seconds() / float64(count)
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
			val, err := s.GetTagDate(t, dateFrom)
			if err != nil {
				val = -1
			}
			resDt[dateFrom] = s.Round(val)
		}
		res[t] = resDt
	}
	return res, nil
}

func (s *BaseStoreImpl) GetTagCountGroup(tag string, from time.Time, to time.Time, count int, group string) (map[string]map[time.Time]float32, error) {
	logger.Debug(fmt.Sprintf("GetTagCount %s : %s - %s (%d)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), count))
	if count == 0 {
		return nil, errors.CountIsEmpty
	}
	if count < 1 {
		return nil, errors.CountIsLessThanOne
	}

	tmDiff := to.Sub(from).Seconds() / float64(count)
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 1; i <= count; i++ {
			dateFrom := from.Add(time.Duration(tmDiff*float64(i-1)) * time.Second)
			dateTo := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
			val, err := s.GetTagFromToGroup(t, dateFrom, dateTo, group)
			if err != nil {
				val = -1
			}
			resDt[dateFrom] = val
		}
		res[t] = resDt
	}
	return res, nil
}

func (s *BaseStoreImpl) GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s : %s - %s", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))
	count := int(to.Sub(from).Seconds())
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(i) * time.Second)
			if val, err := s.cache.Get(tag, dateFrom); err == nil {
				resDt[dateFrom] = s.Round(val)
			} else {
				val, err := s.GetTagDate(t, dateFrom)
				if err != nil {
					val = -1
				}
				resDt[dateFrom] = s.Round(val)
				if val != -1 {
					s.cache.Set(tag, dateFrom, val)
				}
			}
		}
		res[t] = resDt
	}
	return res, nil
}

func (s *BaseStoreImpl) GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s: %s - %s (%s)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), group))

	group = strings.ToLower(group)
	var query string

	switch group {
	case "avg", "sum", "min", "max":
		query = s.config.GetString(fmt.Sprintf("db.%s.query.get_tag_from_to_group", s.config.GetString("app.db.name")))
	case "dif":
		query = s.config.GetString(fmt.Sprintf("db.%s.query.get_tag_from_to_dif", s.config.GetString("app.db.name")))
	case "count":
		query = s.config.GetString(fmt.Sprintf("db.%s.query.get_tag_from_to_count", s.config.GetString("app.db.name")))
	default:
		return -1, errors.GroupError
	}

	fromStr, toStr := from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")

	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr, "{group}": group}, query)

	if query == "" {
		return -1, errors.GroupError
	}

	var value float32
	err := s.db.QueryRow(query).Scan(&value)

	if err != nil {
		return -1, err
	}

	return s.Round(value), nil
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
		logger.Debug(err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var tag string
			err := cur.Scan(&tag)
			if err != nil {
				logger.Debug(err.Error())
				return nil, err
			} else {
				tags["tags"] = append(tags["tags"], tag)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Debug(err.Error())
		}
	}()
	return tags, nil
}

func (s *BaseStoreImpl) GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetDownDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_down_dates")
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.QueryError
	}
	var dates []time.Time
	cur, err := s.db.Query(query)
	if err != nil {
		logger.Debug(err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var date time.Time
			err := cur.Scan(&date)
			if err != nil {
				logger.Debug(err.Error())
				return nil, err
			} else {
				dates = append(dates, date)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Debug(err.Error())
		}
	}()
	return dates, nil
}

func (s *BaseStoreImpl) GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetUpDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("db." + s.config.GetString("app.db.name") + ".query.get_up_dates")
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.QueryError
	}
	var dates []time.Time
	cur, err := s.db.Query(query)
	if err != nil {
		logger.Debug(err.Error())
		return nil, err
	} else {
		for cur.Next() {
			var date time.Time
			err := cur.Scan(&date)
			if err != nil {
				logger.Debug(err.Error())
				return nil, err
			} else {
				dates = append(dates, date)
			}
		}
	}
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Debug(err.Error())
		}
	}()
	return dates, nil
}

func (s *BaseStoreImpl) TemplateGet(name string) (string, error) {
	var body string
	err := s.db.QueryRow("SELECT t.Body from runtime.templates t where t.Name = '" + name + "'").Scan(&body)
	if err != nil {
		return "", err
	}
	return body, nil
}

func (s *BaseStoreImpl) TemplateExec(name string, params map[string]string) (string, error) {
	body, err := s.TemplateGet(name)
	if err != nil {
		return "", err
	}

	for k, v := range params {
		body = strings.Replace(body, "{"+k+"}", v, -1)
	}

	rows, err := s.db.Query(body)
	if err != nil {
		return "", err
	}

	res := ""
	for rows.Next() {
		var line string
		err := rows.Scan(&line)
		if err != nil {
			return "", err
		}
		res += line + "\n"
	}

	return res, nil
}

func (s *BaseStoreImpl) TemplateList(like string) (map[string]string, error) {
	tmpl := map[string]string{}
	if like == "" {
		like = "%"
	}
	q := fmt.Sprintf("SELECT Name, Body FROM runtime.templates WHERE Name LIKE '%s%%'", like)
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name, body string
		err := rows.Scan(&name, &body)
		if err != nil {
			return nil, err
		}
		tmpl[name] = body
	}
	return tmpl, nil
}

func (s *BaseStoreImpl) TemplateSet(name string, body string) error {
	_, err := s.db.Exec("UPDATE runtime.templates SET Body = ? WHERE Name = ?", body, name)
	if err != nil {
		return err
	}
	return nil
}

func (s *BaseStoreImpl) TemplateDel(name string) error {
	_, err := s.db.Exec("DELETE FROM runtime.templates WHERE Name = ?", name)
	if err != nil {
		return err
	}
	return nil
}

func (s *BaseStoreImpl) TemplateAdd(name string, body string) error {
	id := uuid.New().String()
	_, err := s.db.Exec("INSERT INTO runtime.templates (ID, Name, Body) VALUES (?, ?, ?)", id, name, body)
	if err != nil {
		return err
	}
	return nil
}
