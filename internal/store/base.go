package store

// todo: ? rebuild all funcs to return map[string]map[time.Time]float32
// fix: rebuild all funcs to return []map[string]float32
// bug: templates works on clickhouse only! rewrite logic and config for templates

import (
	"database/sql"
	"fmt"
	"net"
	"robin2/internal/errors"
	"strings"
	"time"

	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/data"
	"robin2/internal/logger"
	"robin2/internal/utils"

	"github.com/google/uuid"
)

type Base struct {
	// Store
	db            *sql.DB
	connDriver    string
	connString    string
	config        config.Config
	cache         cache.Cache
	roundConstant float64
	round         int
}

// getMaxConnLimit возвращает лимит максимальных соединений.
// Если MaxConnLimit установлен в конфиге, используется он.
// Иначе используется значение по умолчанию 100.
func (s *Base) getMaxConnLimit() int {
	if s.config.CurrDB.MaxConnLimit > 0 {
		return s.config.CurrDB.MaxConnLimit
	}
	return 100
}

func (s *Base) applyPoolSettings(db *sql.DB) {
	maxConnLimit := s.getMaxConnLimit()
	if s.config.CurrDB.MaxOpenConns > 0 && s.config.CurrDB.MaxOpenConns < maxConnLimit {
		maxConnLimit = s.config.CurrDB.MaxOpenConns
	}
	db.SetMaxOpenConns(maxConnLimit)

	if s.config.CurrDB.MaxIdleConns > 0 {
		db.SetMaxIdleConns(s.config.CurrDB.MaxIdleConns)
	}
	if s.config.CurrDB.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(s.config.CurrDB.ConnMaxIdleTime) * time.Second)
	}
	if s.config.CurrDB.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(s.config.CurrDB.ConnMaxLifetime) * time.Second)
	}
}

func (s *Base) connectWithReuse(name string, cache cache.Cache) error {
	driver := s.config.CurrDB.Type
	connStr := s.GenerateConnectionString(name)
	s.cache = cache

	if s.db != nil && s.connDriver == driver && s.connString == connStr {
		s.applyPoolSettings(s.db)
		if err := s.db.Ping(); err == nil {
			s.logConnection(name)
			return nil
		}
		if err := s.db.Close(); err != nil {
			logger.Error(err.Error())
		}
		s.db = nil
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Error(err.Error())
			return err
		}
		s.db = nil
	}

	db, err := sql.Open(driver, connStr)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	s.applyPoolSettings(db)
	if err = db.Ping(); err != nil {
		db.Close()
		logger.Error(err.Error())
		return err
	}

	s.db = db
	s.connDriver = driver
	s.connString = connStr
	s.logConnection(name)
	return nil
}

// GenerateConnectionString генерирует строку подключения на основе настроек конфигурации.
//
// Он извлекает строку подключения из конфигурации, используя имя базы данных, и заменяет все
// заполнители в строке подключения соответствующими значениями из конфигурации.
// Функция также устанавливает свойство 'round' экземпляра BaseStoreImpl на значение, указанное в
// конфигурации.
//
// Возвращает сгенерированную строку подключения.
func (s *Base) GenerateConnectionString(name string) string {
	connStr := s.config.CurrDB.ConnectionString
	connStr = strings.ReplaceAll(connStr, "{host}", s.config.CurrDB.Host)
	connStr = strings.ReplaceAll(connStr, "{port}", s.config.CurrDB.Port)
	connStr = strings.ReplaceAll(connStr, "{user}", s.config.CurrDB.User)
	connStr = strings.ReplaceAll(connStr, "{password}", s.config.CurrDB.Password)
	connStr = strings.ReplaceAll(connStr, "{database}", s.config.CurrDB.Database)
	s.round = s.config.Round
	return connStr
}

// logConnection записывает соединение с базой данных.
//
// Эта функция извлекает необходимую информацию из конфигурационного файла,
// чтобы записать детали соединения. Она использует имя базы данных, чтобы получить
// хост и порт из конфигурации. Затем она находит IP-адрес хоста и записывает
// детали соединения вместе с полученными IP-адресами.
func (s *Base) logConnection(_ string) {
	dbType := s.config.CurrDB.Type
	host := s.config.CurrDB.Host
	port := s.config.CurrDB.Port
	nips, _ := net.LookupIP(host)
	var ips []string
	for _, ip := range nips {
		ips = append(ips, ip.String())
	}
	logger.Info(fmt.Sprintf("connecting to %s database on %s:%s ( %s )", dbType, host, port, strings.Join(ips, ", ")))
}

// replaceTemplate заменяет все строки в query на соответствующие значения из repMap.
//
// repMap: map[string]string, содержащий все заменяемые значения
func (s *Base) replaceTemplate(repMap map[string]string, query string) string {
	for k, v := range repMap {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

func escapeSQLString(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

func normalizeTags(tag string) []string {
	tags := strings.Split(tag, ",")
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func (s *Base) buildBatchTagQuery(baseQuery string, tags []string) (string, bool) {
	if len(tags) == 0 {
		return "", false
	}

	if len(tags) == 1 {
		return strings.Replace(baseQuery, "{tag}", escapeSQLString(tags[0]), 1), true
	}

	quoted := make([]string, len(tags))
	for i, t := range tags {
		quoted[i] = "'" + escapeSQLString(t) + "'"
	}
	inClause := "(" + strings.Join(quoted, ",") + ")"

	replacements := []struct {
		from string
		to   string
	}{
		{"(h.TagName) = '{tag}'", "(h.TagName) IN " + inClause},
		{"(h.TagName)='{tag}'", "(h.TagName) IN " + inClause},
		{"h.TagName = '{tag}'", "h.TagName IN " + inClause},
		{"h.TagName='{tag}'", "h.TagName IN " + inClause},
		{"(TagName) = '{tag}'", "(TagName) IN " + inClause},
		{"(TagName)='{tag}'", "(TagName) IN " + inClause},
		{"TagName = '{tag}'", "TagName IN " + inClause},
		{"TagName='{tag}'", "TagName IN " + inClause},
	}

	for _, replacement := range replacements {
		if strings.Contains(baseQuery, replacement.from) {
			return strings.Replace(baseQuery, replacement.from, replacement.to, 1), true
		}
	}

	return "", false
}

func (s *Base) buildBatchTagDateQuery(baseQuery string, tags []string, date time.Time) (string, bool) {
	if len(tags) == 0 {
		return "", false
	}

	query := s.replaceTemplate(map[string]string{
		"{date}": date.Format("2006-01-02 15:04:05"),
	}, baseQuery)

	if len(tags) == 1 {
		return strings.Replace(query, "{tag}", escapeSQLString(tags[0]), 1), true
	}

	lowered := strings.ToLower(query)
	if strings.Contains(lowered, " order by ") || strings.Contains(lowered, " limit ") || strings.Contains(lowered, " top ") {
		return "", false
	}

	return s.buildBatchTagQuery(query, tags)
}

func (s *Base) runTagQuery(query string) (data.Tags, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := data.Tags{}
	for rows.Next() {
		var (
			name  string
			date  time.Time
			value sql.NullFloat64
		)
		if err := rows.Scan(&name, &date, &value); err != nil {
			return nil, err
		}
		if !value.Valid {
			continue
		}
		currTag := &data.Tag{
			Name:  name,
			Date:  date,
			Value: float32(value.Float64),
		}
		res = append(res, currTag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetStatus возвращает статус из BaseStoreImpl.
//
// Он возвращает две строки, представляющие версию и время работы,
// а также ошибку, если возникла проблема при получении статуса.
func (s *Base) GetStatus() (string, time.Duration, error) {
	var version string
	var uptime time.Duration
	err := s.db.QueryRow(s.config.CurrDB.Query["status"]).Scan(&version, &uptime)
	if err != nil {
		return "", 0, err
	}
	return version, uptime * time.Second, nil
}

// GetTagDate получает значение, связанное с определенным тегом и датой из хранилища.
//
// Параметры:
// - tag: тег, для которого нужно получить значение.
// - date: дата, для которой нужно получить значение.
//
// Возвращает:
// - *data.Tag: значение, связанное с определенным тегом и датой.
// - error: любая ошибка, возникшая в процессе получения значения.
func (s *Base) GetTagDate(tag string, date time.Time) (*data.Tag, error) {
	if date.IsZero() {
		return nil, errors.ErrInvalidDate
	}
	if s.db == nil {
		return nil, errors.ErrDbConnectionFailed
	}

	currTag := data.Tag{
		Name:  tag,
		Date:  date,
		Value: -1,
	}

	// day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	// _, err := s.cache.Get(tag+"|"+day.Format("2006-01-02"), day)
	// if err != nil {
	// 	s.cacheDay(tag, day)
	// }

	if val, err := s.cache.Get(tag, date); err == nil {
		currTag.Value = val
		return &currTag, nil
	}

	query := s.config.CurrDB.Query["get_tag_date"]
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{date}": date.Format("2006-01-02 15:04:05")}, query)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}

	switch len(cols) {
	case 1:
		var value sql.NullFloat64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		if value.Valid {
			currTag.Value = float32(value.Float64)
		}
	case 3:
		var (
			name  sql.NullString
			ts    sql.NullTime
			value sql.NullFloat64
		)
		if err := rows.Scan(&name, &ts, &value); err != nil {
			return nil, err
		}
		if name.Valid && name.String != "" {
			currTag.Name = name.String
		}
		if ts.Valid {
			currTag.Date = ts.Time
		}
		if value.Valid {
			currTag.Value = float32(value.Float64)
		}
	default:
		return nil, fmt.Errorf("unsupported get_tag_date column count: %d", len(cols))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if s.cache != nil && currTag.Value != -1 {
		err = s.cache.Set(currTag.Name, date, currTag.Value)
		if err != nil {
			logger.Error(err.Error())
		}
	}

	return &currTag, nil
}

func (s *Base) GetTagsDate(tags []string, date time.Time) (data.Tags, error) {
	if date.IsZero() {
		return nil, errors.ErrInvalidDate
	}
	if s.db == nil {
		return nil, errors.ErrDbConnectionFailed
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	if len(normalized) == 0 {
		return data.Tags{}, nil
	}

	found := make(map[string]*data.Tag, len(normalized))
	missing := make([]string, 0, len(normalized))
	for _, tag := range normalized {
		if val, err := s.cache.Get(tag, date); err == nil {
			found[tag] = &data.Tag{Name: tag, Date: date, Value: val}
			continue
		}
		missing = append(missing, tag)
	}

	if len(missing) > 0 {
		baseQuery := s.config.CurrDB.Query["get_tag_date"]
		batchResult, batchErr := data.Tags(nil), error(nil)
		if query, ok := s.buildBatchTagDateQuery(baseQuery, missing, date); ok {
			batchResult, batchErr = s.runTagQuery(query)
		}

		if batchErr == nil && len(batchResult) > 0 {
			for _, tagValue := range batchResult {
				found[tagValue.Name] = tagValue
				if s.cache != nil && tagValue.Value != -1 {
					if err := s.cache.Set(tagValue.Name, date, float32(tagValue.Value)); err != nil {
						logger.Error(err.Error())
					}
				}
			}
		}

		for _, tag := range missing {
			if _, ok := found[tag]; ok {
				continue
			}
			tagValue, err := s.GetTagDate(tag, date)
			if err != nil {
				continue
			}
			found[tag] = tagValue
		}
	}

	res := make(data.Tags, 0, len(normalized))
	for _, tag := range normalized {
		if tagValue, ok := found[tag]; ok {
			res = append(res, tagValue)
		}
	}
	return res, nil
}

// func (s *Base) cacheDay(tag string, day time.Time) {
// 	to := day.AddDate(0, 0, 1)
// 	s.GetTagFromToUncached(tag, day, to)
// 	s.cache.Set(tag+"|"+day.Format("2006-01-02"), day, 1)
// }

// GetTagCount вычисляет значение указанного тега в заданном количестве внутри временного диапазона.
//
// Параметры:
// - tag: Тег, который нужно посчитать.
// - from: Начальное время диапазона.
// - to: Конечное время диапазона.
// - count: Количество интервалов внутри диапазона.
//
// Возвращает:
// - map[string]map[time.Time]float32: Карта, содержащая количество тегов для каждого тега и временного интервала.
// - error: Ошибка, если количество равно нулю или меньше единицы.
func (s *Base) GetTagCount(tag string, from time.Time, to time.Time, count int) (map[string]map[time.Time]float32, error) {
	logger.Debug(fmt.Sprintf("GetTagCount %s : %s - %s (%d)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), count))
	if count == 0 {
		return nil, errors.ErrCountIsEmpty
	}
	if count < 1 {
		return nil, errors.ErrCountIsLessThanOne
	}

	tmDiff := to.Sub(from).Seconds() / float64(count)
	tags := strings.Split(tag, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
			valOut, err := s.GetTagDate(t, dateFrom)
			if err != nil {
				return nil, err
			}
			val := valOut.Value
			resDt[dateFrom] = float32(val)
		}
		res[t] = resDt
	}
	return res, nil
}

// GetTagCountGroup получает значения тегов, в нужном количестве, сгруппированных по интервалам времени.
//
// Принимает следующие параметры:
// - tag: тег, для которого нужно получить количество.
// - from: начальное время интервала.
// - to: конечное время интервала.
// - count: количество интервалов для разделения временного диапазона.
// - group: группа для классификации результатов.
//
// Возвращает:
// - data.Tags: слайс с результатами.
// - error: в случае, если количество равно нулю или меньше единицы,
func (s *Base) GetTagCountGroup(tag string, from time.Time, to time.Time, count int, group string) (data.Tags, error) {
	logger.Debug(fmt.Sprintf("GetTagCount %s : %s - %s (%d)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), count))
	if count == 0 {
		return nil, errors.ErrCountIsEmpty
	}
	if count < 1 {
		return nil, errors.ErrCountIsLessThanOne
	}

	tmDiff := to.Sub(from).Seconds() / float64(count)
	tags := strings.Split(tag, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	res := data.Tags{}

	if group == "avgm" {
		for _, t := range tags {
			allPeriod := data.Tags{}
			for i := 1; i <= count; i++ {

				dateFrom := from.Add(time.Duration(tmDiff*float64(i-1)) * time.Second)
				dateTo := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
				fromStr := dateFrom.Format("2006-01-02 15:04:05")
				toStr := dateTo.Format("2006-01-02 15:04:05")

				val, err := s.cache.GetStr(t, fromStr+"|"+toStr+"|"+group)
				if err != nil {
					if len(allPeriod) == 0 {
						allPeriod, err = s.GetTagFromTo(t, from, to)
						if err != nil {
							return nil, err
						}
					}
					val = allPeriod.GetFromTo(dateFrom, dateTo).Average(t)
					if val != -1 {
						err := s.cache.SetStr(t, fromStr+"|"+toStr+"|"+group, val)
						if err != nil {
							logger.Error(err.Error())
						}
					}
				}
				resDt := data.Tag{
					Name:  t,
					Date:  dateTo,
					Value: val,
				}
				res = append(res, &resDt)
			}
		}
	} else {
		for _, t := range tags {
			for i := 1; i <= count; i++ {
				dateFrom := from.Add(time.Duration(tmDiff*float64(i-1)) * time.Second)
				dateTo := from.Add(time.Duration(tmDiff*float64(i)) * time.Second)
				val, err := s.GetTagFromToGroup(t, dateFrom, dateTo, group)
				if err != nil {
					val = -1
				}
				resDt := data.Tag{
					Name:  t,
					Date:  dateTo,
					Value: val,
				}
				res = append(res, &resDt)
			}
		}
	}
	return res, nil
}

// GetTagFromTo извлекает данные для указанного тега в заданном временном диапазоне.
//
// Параметры:
// - tag: Тег, для которого нужно извлечь данные.
// - from: Начальное время временного диапазона.
// - to: Конечное время временного диапазона.
//
// Возвращает:
// - data.Tags: Список извлеченных данных.
// - error: Ошибка, если данные не были получены.
// func (s *Base) GetTagFromTo(tag string, from time.Time, to time.Time) (data.Tags, error) {
// 	logger.Debug(fmt.Sprintf("GetTagFromTo %s : %s - %s", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

// 	tags := strings.Split(tag, ",")
// 	for i, t := range tags {
// 		tags[i] = strings.TrimSpace(t)
// 	}

// 	res := data.Tags{}

// 	for _, t := range tags {

// 		query := s.replaceTemplate(map[string]string{
// 			"{tag}":  t,
// 			"{from}": from.Format("2006-01-02 15:04:05"),
// 			"{to}":   to.Format("2006-01-02 15:04:05"),
// 		}, s.config.CurrDB.Query["get_tag_from_to"])

// 		rows, err := s.db.Query(query)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for rows.Next() {
// 			var tag string
// 			var date time.Time
// 			var val float32
// 			err = rows.Scan(&tag, &date, &val)
// 			if err != nil {
// 				return nil, err
// 			}

// 			// for i := 1; i <= int(to.Sub(from).Seconds()); i++ {
// 			// 	date := from.Add(time.Duration(i) * time.Second)
// 			// 	currTag, err := s.GetTagDate(t, date)
// 			// 	if err != nil {
// 			// 		return nil, err
// 			// 	}
// 			// 	res = append(res, currTag)
// 			// }
// 			currTag := &data.Tag{
// 				Name:  t,
// 				Date:  date,
// 				Value: val,
// 			}
// 			res = append(res, currTag)
// 		}
// 		rows.Close()
// 	}
// 	return res, nil
// }

func (s *Base) GetTagFromTo(tag string, from time.Time, to time.Time) (data.Tags, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s : %s - %s", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

	tags := normalizeTags(tag)
	if len(tags) == 0 {
		return data.Tags{}, nil
	}

	baseQuery := s.replaceTemplate(map[string]string{
		"{from}": from.Format("2006-01-02 15:04:05"),
		"{to}":   to.Format("2006-01-02 15:04:05"),
	}, s.config.CurrDB.Query["get_tag_from_to"])

	if query, ok := s.buildBatchTagQuery(baseQuery, tags); ok {
		return s.runTagQuery(query)
	}

	res := data.Tags{}
	for _, t := range tags {
		query := s.replaceTemplate(map[string]string{"{tag}": escapeSQLString(t)}, baseQuery)
		tagValues, err := s.runTagQuery(query)
		if err != nil {
			return nil, err
		}
		res = append(res, tagValues...)
	}

	return res, nil
}

func (s *Base) GetTagFromToUncached(tag string, from time.Time, to time.Time) (data.Tags, error) {
	//	logger.Debug(fmt.Sprintf("GetTagFromToUncached %s : %s - %s", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

	tags := strings.Split(tag, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}

	// res := data.Tags{}

	for _, t := range tags {

		query := s.replaceTemplate(map[string]string{
			"{tag}":  t,
			"{from}": from.Format("2006-01-02 15:04:05"),
			"{to}":   to.Format("2006-01-02 15:04:05"),
		}, s.config.CurrDB.Query["get_tag_from_to"])

		rows, err := s.db.Query(query)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var tag string
			var date time.Time
			var val float32
			err = rows.Scan(&tag, &date, &val)
			if err != nil {
				continue
			}
			// currTag := data.Tag{
			// 	Name:  tag,
			// 	Date:  date,
			// 	Value: val,
			// }
			err = s.cache.Set(tag, date, val)
			if err != nil {
				logger.Error(err.Error())
			}
			// res = append(res, &currTag)
		}

		rows.Close()
	}

	return nil, nil
}

// GetTagFromToGroup извлекает значение типа float32 для указанного тега в заданном временном диапазоне и группе.
//
// Параметры:
// - tag: Тег, для которого нужно извлечь значение.
// - from: Начальное время временного диапазона.
// - to: Конечное время временного диапазона.
// - group: Метод группировки, такой как "avg", "sum", "min", "max", "dif" или "count".
//
// Возвращает:
// - float32: Извлеченное значение.
// - error: Ошибка, если извлечение не удалось.
func (s *Base) GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s: %s - %s (%s)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), group))

	group = strings.ToLower(group)
	var query string

	fromStr, toStr := from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")
	if val, err := s.cache.GetStr(tag, fromStr+"|"+toStr+"|"+group); err == nil {
		return val, nil
	}

	switch group {
	case "avg", "sum", "min", "max":
		query = s.config.CurrDB.Query["get_tag_from_to_group_fast"]
		if query == "" {
			query = s.config.CurrDB.Query["get_tag_from_to_group"]
		}

	case "dif":
		query = s.config.CurrDB.Query["get_tag_from_to_group_dif"]

	case "count":
		query = s.config.CurrDB.Query["get_tag_from_to_group_count"]

	case "avgm":
		t, err := s.GetTagFromTo(tag, from, to)
		if err != nil {
			return -1, err
		}
		val := t.Average(tag)
		err = s.cache.SetStr(tag, fromStr+"|"+toStr+"|"+group, val)
		if err != nil {
			logger.Error(err.Error())
		}
		return val, nil

	default:
		return -1, errors.ErrGroupError
	}

	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr, "{group}": group}, query)

	if query == "" {
		return -1, errors.ErrGroupError
	}

	var value sql.NullFloat64
	row := s.db.QueryRow(query)
	err := row.Scan(&value)

	if err != nil {
		return -1, err
	}

	if !value.Valid {
		return -1, nil
	}

	err = s.cache.SetStr(tag, fromStr+"|"+toStr+"|"+group, float32(value.Float64))
	if err != nil {
		logger.Error(err.Error())
	}

	return float32(value.Float64), nil
}

// GetTagList извлекает список тегов, соответствующих заданному шаблону.
//
// Параметры:
// - like: Шаблон для сопоставления с тегами.
//
// Возвращает:
// - *data.Output: Список тегов.
// - error: Ошибка, если запрос к базе данных не выполнен.
func (s *Base) GetTagList(like string) (*data.Output, error) {
	if like == "" {
		like = "%"
	}
	like = s.replaceTemplate(map[string]string{"*": "%", "?": "_", " ": "%"}, like)
	query := s.config.CurrDB.Query["get_tag_list"]
	// replace {tag} with like
	query = strings.Replace(query, "{tag}", like, -1)
	// tags := make([]string, 0, 15000)
	rows, err := s.db.Query(query)
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Debug(err.Error())
		}
	}()
	if err != nil {
		logger.Debug(err.Error())
		return nil, err
	}

	out := &data.Output{}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out.Headers = append(out.Headers, cols...)

	row := make([]interface{}, len(cols))
	for i := range row {
		row[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		err = rows.Scan(row...)
		if err != nil {
			return nil, err
		}

		values := make([]string, len(cols))
		for i, v := range row {
			values[i] = string(*v.(*sql.RawBytes))
		}

		out.Rows = append(out.Rows, values)
	}

	return out, nil
}

// GetDownDates получает список дат отключений в указанном временном диапазоне, отфильтрованный по тегу.
//
// Параметры:
// - tag: строка, представляющая тег для фильтрации дат отключений.
// - from: time.Time, представляющий начало временного диапазона.
// - to: time.Time, представляющий конец временного диапазона.
//
// Возвращает:
// - []time.Time: срез time.Time, представляющий даты отключений в указанном диапазоне.
// - error: ошибка, если возникла в процессе получения данных.
func (s *Base) GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetDownDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.CurrDB.Query["get_down_dates"]
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.ErrQueryError
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

// GetUpDates возвращает список дат включений в указанном временном диапазоне, отфильтрованный по тегу.
//
// Параметры:
// - tag: Тег, используемый для фильтрации объектов time.Time.
// - from: Начальное время для фильтрации объектов time.Time.
// - to: Конечное время для фильтрации объектов time.Time.
//
// Возвращает:
// - []time.Time: Список объектов time.Time, удовлетворяющих заданным критериям.
// - error: Объект ошибки, если возникла проблема при получении объектов time.Time.
func (s *Base) GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetUpDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.CurrDB.Query["get_up_dates"]
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr}, query)
	if query == "" {
		return nil, errors.ErrQueryError
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

// TemplateGet получает тело шаблона по его имени.
//
// Параметры:
// - name: имя шаблона.
//
// Возвращает:
// - string: тело шаблона.
// - error: ошибка, если шаблон не может быть найден или происходит ошибка при получении.
func (s *Base) TemplateGet(name string) (string, error) {
	var body string
	err := s.db.QueryRow("SELECT t.Body from runtime.templates t where t.Name = '" + name + "'").Scan(&body)
	if err != nil {
		return "", err
	}
	return body, nil
}

// TemplateExec выполняет шаблон с заданным именем и параметрами.
//
// Он заменяет заполнители в теле шаблона значениями из карты params и затем
// выполняет полученный SQL-запрос с помощью базового подключения к базе данных.
// Возвращает результат в виде строки.
//
// Параметры:
//   - name: имя шаблона для выполнения.
//   - params: карта пар ключ-значение, представляющая параметры для замены
//     в теле шаблона.
//
// Возвращает:
// - string: результат выполнения шаблона.
// - error: ошибка, если произошла во время выполнения.
func (s *Base) TemplateExec(name string, params map[string]string) (*data.Output, error) {
	body, err := s.TemplateGet(name)
	if err != nil {
		return nil, err
	}

	for k, v := range params {
		body = strings.Replace(body, "{"+k+"}", v, -1)
	}

	// todo: add cache

	dbName := utils.ThenIf(params["db"] != "", params["db"], s.config.CurrDB.Name)
	// var storedb *Store
	// if dbName != s.config.CurrDB.Name {
	storedb, err := New(s.config)
	if err != nil {
		return nil, err
	}
	err = storedb.Connect(dbName, nil)
	if err != nil {
		return nil, err
	}
	// } else {
	// 	storedb = &s
	// }

	rows, err := storedb.ExecQuery(body)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// TemplateList получает карту имен и тел шаблонов на основе заданного шаблона.
//
// Параметр `like` используется для указания шаблона для сопоставления имен шаблонов. Если `like` является пустой строкой, шаблон по умолчанию устанавливается на "%".
// Функция возвращает карту map[string]string, содержащую имена шаблонов в качестве ключей и их тела в качестве значений.
// Если произошла ошибка при выполнении запроса к базе данных, функция возвращает nil и ошибку.
func (s *Base) TemplateList(like string) (map[string]string, error) {
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

// TemplateSet обновляет тело шаблона с указанным именем в таблице runtime.templates.
//
// Параметры:
// - name: имя шаблона.
// - body: новое тело шаблона.
// Возвращает:
// - error: если произошла ошибка при обновлении шаблона.
func (s *Base) TemplateSet(name string, body string) error {
	_, err := s.db.Exec("UPDATE runtime.templates SET Body = ? WHERE Name = ?", body, name)
	if err != nil {
		return err
	}
	return nil
}

// TemplateDel удаляет шаблон по имени из BaseStoreImpl.
//
// name - Имя шаблона, который должен быть удален.
// error - Возвращает ошибку, если удаление не удалось.
func (s *Base) TemplateDel(name string) error {
	// _, err := s.db.Exec("DELETE FROM runtime.templates WHERE Name = ?", name)
	_, err := s.db.Exec("ALTER TABLE runtime.templates DELETE WHERE Name = ?", name)
	if err != nil {
		return err
	}
	return nil
}

// TemplateAdd добавляет новый шаблон в BaseStoreImpl.
//
// Принимает два параметра:
// - name: строка, представляющая имя шаблона.
// - body: строка, представляющая тело шаблона.
//
// Возвращает ошибку, указывающую на любые проблемы, возникшие во время операции.
func (s *Base) TemplateAdd(name string, body string) error {
	id := uuid.New().String()
	_, err := s.db.Exec("INSERT INTO runtime.templates (ID, Name, Body) VALUES (?, ?, ?)", id, name, body)
	if err != nil {
		return err
	}
	return nil
}

func (s *Base) ExecQuery(query string) (*data.Output, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &data.Output{}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out.Headers = append(out.Headers, cols...)

	row := make([]interface{}, len(cols))
	for i := range row {
		row[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		err = rows.Scan(row...)
		if err != nil {
			return nil, err
		}

		// Convert row to []string
		strRow := make([]string, len(row))
		for i, v := range row {
			strRow[i] = string(*v.(*sql.RawBytes))
		}

		out.Rows = append(out.Rows, strRow)
		out.Count = len(out.Rows)
	}

	return out, nil
}
