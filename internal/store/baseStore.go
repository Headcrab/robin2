package store

import (
	"database/sql"
	"fmt"
	"net"
	"robin2/internal/errors"
	"strings"
	"time"

	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"

	"github.com/google/uuid"
	// _ "github.com/denisenkom/go-mssqldb"
	// _ "github.com/go-sql-driver/mysql"
)

/*------------------------------------------------------------------*/

type BaseStore interface {
	Connect(name string, cache cache.BaseCache) error
	GetTagDate(tag string, date time.Time) (float32, error)
	GetTagCount(tag string, from time.Time, to time.Time, strCount int) (map[string]map[time.Time]float32, error)
	GetTagCountGroup(tag string, from time.Time, to time.Time, strCount int, group string) (map[string]map[time.Time]float32, error)
	GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error)
	GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error)
	GetTagList(like string) ([]string, error)
	GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error)
	GetStatus() (string, string, error)

	TemplateList(like string) (map[string]string, error)
	TemplateExec(name string, params map[string]string) ([]map[string]string, error)

	TemplateAdd(name string, body string) error
	TemplateSet(name string, body string) error
	TemplateGet(name string) (string, error)
	TemplateDel(name string) error

	ExecQuery(query string) ([]map[string]string, error)
}

type BaseStoreImpl struct {
	BaseStore
	db            *sql.DB
	config        config.Config
	cache         cache.BaseCache
	roundConstant float64
	round         int
}

func thenIf[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// marshalConnectionString генерирует строку подключения на основе настроек конфигурации.
//
// Он извлекает строку подключения из конфигурации, используя имя базы данных, и заменяет все
// заполнители в строке подключения соответствующими значениями из конфигурации.
// Функция также устанавливает свойство 'round' экземпляра BaseStoreImpl на значение, указанное в
// конфигурации.
//
// Возвращает сгенерированную строку подключения.
func (s *BaseStoreImpl) marshalConnectionString(name string) string {
	connStr := s.config.GetString("app.db." + name + ".connection_string")
	for k, v := range s.config.GetStringMapString("app.db." + name) {
		connStr = strings.ReplaceAll(connStr, "{"+k+"}", v)
	}
	s.round = s.config.GetInt("app.round")
	return connStr
}

// logConnection записывает соединение с базой данных.
//
// Эта функция извлекает необходимую информацию из конфигурационного файла,
// чтобы записать детали соединения. Она использует имя базы данных, чтобы получить
// хост и порт из конфигурации. Затем она находит IP-адрес хоста и записывает
// детали соединения вместе с полученными IP-адресами.
func (s *BaseStoreImpl) logConnection(dbName string) {
	dbType := s.config.GetString("app.db." + dbName + ".type")
	host := s.config.GetString("app.db." + dbName + ".host")
	port := s.config.GetString("app.db." + dbName + ".port")
	nips, _ := net.LookupIP(host)
	var ips []string
	for _, ip := range nips {
		ips = append(ips, ip.String())
	}
	logger.Info(fmt.Sprintf("connected to %s database on %s:%s ( %s )", dbType, host, port, strings.Join(ips, ", ")))
}

// replaceTemplate заменяет все строки в query на соответствующие значения из repMap.
//
// repMap: map[string]string, содержащий все заменяемые значения
func (s *BaseStoreImpl) replaceTemplate(repMap map[string]string, query string) string {
	for k, v := range repMap {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

// GetStatus возвращает статус из BaseStoreImpl.
//
// Он возвращает две строки, представляющие версию и время работы,
// а также ошибку, если возникла проблема при получении статуса.
func (s *BaseStoreImpl) GetStatus() (string, string, error) {
	var version, uptime string
	err := s.db.QueryRow(s.config.GetString("app.db."+s.config.GetString("app.db.current")+".query.status")).Scan(&version, &uptime)
	if err != nil {
		return "", "", err
	}
	return version, uptime, nil
}

// GetTagDate получает значение, связанное с определенным тегом и датой из хранилища.
//
// Параметры:
// - tag: тег, для которого нужно получить значение.
// - date: дата, для которой нужно получить значение.
//
// Возвращает:
// - float32: значение, связанное с тегом и датой.
// - error: любая ошибка, возникшая в процессе получения значения.
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
	query := s.config.GetString("app.db." + s.config.GetString("app.db.current") + ".query.get_tag_date")
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
			resDt[dateFrom] = val
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
// Возвращает map[string]map[time.Time]float32, представляющую количество тегов, сгруппированных по интервалам времени,
// и ошибку, если она есть.
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

// GetTagFromTo извлекает данные для указанного тега в заданном временном диапазоне.
//
// Параметры:
// - tag: Тег, для которого нужно извлечь данные.
// - from: Начальное время временного диапазона.
// - to: Конечное время временного диапазона.
//
// Возвращает:
// - Карта значений тега для каждого временного штампа в указанном диапазоне.
// - Ошибку, если возникла проблема при извлечении данных.
func (s *BaseStoreImpl) GetTagFromTo(tag string, from time.Time, to time.Time) (map[string]map[time.Time]float32, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s : %s - %s", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))
	count := int(to.Sub(from).Seconds())
	tags := strings.Split(tag, ",")
	res := make(map[string]map[time.Time]float32, len(tags))
	for _, t := range tags {
		resDt := make(map[time.Time]float32, count)
		for i := 0; i < count; i++ {
			dateFrom := from.Add(time.Duration(i) * time.Second)
			// if val, err := s.cache.Get(tag, dateFrom); err == nil {
			// 	resDt[dateFrom] = s.Round(val)
			// } else {
			val, err := s.GetTagDate(t, dateFrom)
			if err != nil {
				val = -1
			}
			resDt[dateFrom] = val
			// if val != -1 {
			// 	s.cache.Set(tag, dateFrom, val)
			// }
			// }
		}
		res[t] = resDt
	}
	return res, nil
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
func (s *BaseStoreImpl) GetTagFromToGroup(tag string, from time.Time, to time.Time, group string) (float32, error) {
	logger.Debug(fmt.Sprintf("GetTagFromTo %s: %s - %s (%s)", tag, from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"), group))

	group = strings.ToLower(group)
	var query string

	switch group {
	case "avg", "sum", "min", "max":
		query = s.config.GetString(fmt.Sprintf("app.db.%s.query.get_tag_from_to_group", s.config.GetString("app.db.current")))
	case "dif":
		query = s.config.GetString(fmt.Sprintf("app.db.%s.query.get_tag_from_to_dif", s.config.GetString("app.db.current")))
	case "count":
		query = s.config.GetString(fmt.Sprintf("app.db.%s.query.get_tag_from_to_count", s.config.GetString("app.db.current")))
	default:
		return -1, errors.GroupError
	}

	fromStr, toStr := from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")

	// check cache

	if val, err := s.cache.GetStr(tag, fromStr+"|"+toStr+"|"+group); err == nil {
		return val, nil
	}

	query = s.replaceTemplate(map[string]string{"{tag}": tag, "{from}": fromStr, "{to}": toStr, "{group}": group}, query)

	if query == "" {
		return -1, errors.GroupError
	}

	var value float32
	err := s.db.QueryRow(query).Scan(&value)

	if err != nil {
		return -1, err
	}

	s.cache.SetStr(tag, fromStr+"|"+toStr+"|"+group, value)

	return value, nil
}

// GetTagList извлекает список тегов, соответствующих заданному шаблону.
//
// Параметры:
// - like: Шаблон для сопоставления с тегами.
//
// Возвращает:
// - map[string][]string: Карта, содержащая теги, сгруппированные по категориям.
// - error: Ошибка, если запрос к базе данных не выполнен.
func (s *BaseStoreImpl) GetTagList(like string) ([]string, error) {
	if like == "" {
		like = "%"
	}
	like = s.replaceTemplate(map[string]string{"*": "%", "?": "_", " ": "%"}, like)
	query := s.config.GetString("app.db." + s.config.GetString("app.db.current") + ".query.get_tag_list")
	// replace {tag} with like
	query = strings.Replace(query, "{tag}", like, -1)
	tags := make([]string, 0, 15000)
	cur, err := s.db.Query(query)
	defer func() {
		err := cur.Close()
		if err != nil {
			logger.Debug(err.Error())
		}
	}()
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
				tags = append(tags, tag)
			}
		}
	}
	return tags, nil
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
func (s *BaseStoreImpl) GetDownDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetDownDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("app.db." + s.config.GetString("app.db.current") + ".query.get_down_dates")
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
func (s *BaseStoreImpl) GetUpDates(tag string, from time.Time, to time.Time) ([]time.Time, error) {
	logger.Debug("GetUpDate " + tag + " : " + from.Format("2006-01-02 15:04:05") + " - " + to.Format("2006-01-02 15:04:05"))
	var query string
	query = s.config.GetString("app.db." + s.config.GetString("app.db.current") + ".query.get_up_dates")
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

// TemplateGet получает тело шаблона по его имени.
//
// Параметры:
// - name: имя шаблона.
//
// Возвращает:
// - string: тело шаблона.
// - error: ошибка, если шаблон не может быть найден или происходит ошибка при получении.
func (s *BaseStoreImpl) TemplateGet(name string) (string, error) {
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
func (s *BaseStoreImpl) TemplateExec(name string, params map[string]string) ([]map[string]string, error) {
	body, err := s.TemplateGet(name)
	if err != nil {
		return nil, err
	}

	dbName := thenIf(params["db"] != "", params["db"], s.config.GetString("app.db.current"))

	for k, v := range params {
		body = strings.Replace(body, "{"+k+"}", v, -1)
	}

	storedb := NewFactory().NewStore(s.config.GetString("app.db."+dbName+".type"), s.config)
	if storedb == nil {
		return nil, errors.StoreError
	}
	err = storedb.Connect(dbName, nil)
	if err != nil {
		return nil, err
	}

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

// TemplateSet обновляет тело шаблона с указанным именем в таблице runtime.templates.
//
// Параметры:
// - name: имя шаблона.
// - body: новое тело шаблона.
// Возвращает:
// - error: если произошла ошибка при обновлении шаблона.
func (s *BaseStoreImpl) TemplateSet(name string, body string) error {
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
func (s *BaseStoreImpl) TemplateDel(name string) error {
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
func (s *BaseStoreImpl) TemplateAdd(name string, body string) error {
	id := uuid.New().String()
	_, err := s.db.Exec("INSERT INTO runtime.templates (ID, Name, Body) VALUES (?, ?, ?)", id, name, body)
	if err != nil {
		return err
	}
	return nil
}

func (s *BaseStoreImpl) ExecQuery(query string) ([]map[string]string, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(sql.RawBytes)
	}

	lines := []map[string]string{}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, err
		}

		m := make(map[string]string)
		for i, col := range cols {
			m[col] = string(*vals[i].(*sql.RawBytes))
			// lines += fmt.Sprintf("%s: %s\n", col, vals[i].(*sql.RawBytes))
		}
		lines = append(lines, m)
	}
	return lines, nil
}
