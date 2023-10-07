package robin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"robin2/pkg/logger"
	"strconv"
	"time"
)

// @Summary Получить лог
// @Description Возвращает логи приложения
// @Tags System
// @Produce json
// @Success 200 {array} string
// @Router /api/log/ [get]
func (a *App) handleAPIGetLog(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered log page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	logs, err := logger.GetLogHistory()
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	tagValue, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	w.Write(tagValue)
}

// @Summary Получить значение тега
// @Description Возвращает значение тега на выбранную дату.
// @Tags Tag
// @Produce plain/text json
// @Success 200 {array} string
// @Router /get/tag/ [get]
// @Param tag query string true "Наименование тега"
// @Param date query string false "Дата" "date-time"
// @Param from query string false "Дата начала периода"
// @Param to query string false "Дата окончания периода"
// @Param group query string false "Функция группировки (avg, sum, count, min, max)"
// @Param count query string false "Количество значений"
// @Param round query string false "Округление, знаков после запятой (по умолчанию 2)"
// @Param format query string false "Формат вывода (raw - без округления и замены точки на запятую, только для одного знчения)"
func (a *App) handleAPIGetTag(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	headers.Set("Content-Type", "text/plain")
	procTimeBegin := time.Now()
	writer := []byte("")
	defer func() {
		headers.Set("Procession-Time", time.Since(procTimeBegin).String())
		w.WriteHeader(http.StatusOK)
		w.Write(writer)
	}()

	a.opCount++

	query := r.URL.Query()
	tag := query.Get("tag")
	date := query.Get("date")
	from := query.Get("from")
	to := query.Get("to")
	group := query.Get("group")
	round := query.Get("round")
	count := query.Get("count")
	format := query.Get("format")

	if round != "" {
		round, _ := strconv.Atoi(round)
		a.store.SetRound(round)
	} else {
		round := a.config.GetInt("app.round")
		a.store.SetRound(round)
	}

	if tag != "" && date != "" {
		writer = a.getTagOnDate(tag, date, format)
		return
	}

	if tag != "" && from != "" && to != "" {
		if count != "" {
			if group == "" {
				writer = a.getTagByCount(tag, from, to, count)
				return
			} else {
				writer = a.getTagFromToByCountWithGroup(tag, from, to, count, group)
				return
			}
		}

		if group == "" {
			writer = a.getTagFromTo(tag, from, to)
			return
		} else {
			writer = a.getTagFromToWithGroup(tag, from, to, group, format)
			return
		}
	}
}

// @Summary Получить список тегов
// @Description Возвращает список всех тегов по маске
// @Tags Tag
// @Produce json
// @Success 200 {array} string
// @Router /get/tag/list/ [get]
// @Param like query string false "Маска поиска"
// returns JSON with tags by mask
func (a *App) handleAPIGetTagList(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	procTimeBegin := time.Now()
	writer := []byte("")
	defer func() {
		w.Header().Set("Procession-Time", time.Since(procTimeBegin).String())
		w.WriteHeader(http.StatusOK)
		w.Write(writer)
	}()

	// Extract query parameter
	like := r.URL.Query().Get("like")

	// Retrieve list of tags
	tags, err := a.store.GetTagList(like)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	// Marshal tags into JSON string
	writer, err = json.MarshalIndent(tags, "", "  ")
	if err != nil {
		// logger.Debug(err.error())
		writer = []byte("#Error: " + err.Error())
		return
	}

	// Write JSON string to response
	// w.Write(j)
}

// @Summary Получить даты отключения оборудования
// @Description Возвращает дату и время отключения оборудования
// @Tags Tag
// @Produce plain/text
// @Success 200 {array} string
// @Router /get/tag/down/ [get]
// @Param tag query string true "Наименование тега"
// @Param from query string false "Дата начала периода"
// @Param to query string false "Дата окончания периода"
// @Param count query string false "Номер отключения после даты начала (0 - первое отключение)"
func (a *App) handleAPIGetTagDown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		w.Write([]byte("#Error: tag is empty"))
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 0
	}
	fromT, err := a.excelTimeToTime(from)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	toT, err := a.excelTimeToTime(to)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	v, err := a.store.GetDownDates(tag, fromT, toT)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	} else {
		// tagValue, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		if (count >= 0) && (count < len(v)) {
			val := v[count].Format("2006-01-02 15:04:05")
			w.Write([]byte(val))
		} else {
			w.Write([]byte(""))
		}
	}
}

// @Summary Получить даты включения оборудования
// @Description Возвращает дату и время включения оборудования
// @Tags Tag
// @Produce plain/text
// @Success 200 {array} string
// @Router /get/tag/up/ [get]
// @Param tag query string true "Наименование тега"
// @Param from query string false "Дата начала периода"
// @Param to query string false "Дата окончания периода"
// @Param count query string false "Номер включения после даты начала (0 - первое включение)"
func (a *App) handleAPIGetTagUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	tag := r.URL.Query().Get("tag")
	if tag == "" {
		w.Write([]byte("#Error: tag is empty"))
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 0
	}

	fromT, err := a.excelTimeToTime(from)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	toT, err := a.excelTimeToTime(to)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	v, err := a.store.GetUpDates(tag, fromT, toT)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	if count >= 0 && count < len(v) {
		val := v[count].Format("2006-01-02 15:04:05")
		w.Write([]byte(val))
	} else {
		w.Write([]byte(""))
	}
}

// @Summary Полчить информацию
// @Description Возвращает информацию о приложении
// @Tags System
// @Produce json
// @Success 200 {array} string
// @Router /api/info/ [get]
func (a *App) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered info page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	inf := map[string]interface{}{
		"name":     a.name,
		"version":  a.version,
		"uptime":   time.Since(a.startTime).Round(time.Second).String(),
		"op_count": a.opCount,
	}

	if err := json.NewEncoder(w).Encode(inf); err != nil {
		logger.Error(err.Error())
	}
}

func (a *App) handleAPIReloadConfig(w http.ResponseWriter, r *http.Request) {
	logger.Trace("reloading config file")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	logger.Info("reloading config file")

	if err := a.config.Reload(); err != nil {
		logger.Fatal("Failed to read config file")
	}
}

// @Summary Получить время работы
// @Description Возвращает время работы приложения с времени запуска
// @Tags System
// @Produce plain/text
// @Success 200 {array} string
// @Router /api/uptime/ [get]
func (a *App) handleAPIUptime(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered uptime page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	uptime := time.Since(a.startTime).Round(time.Second)
	_, err := w.Write([]byte(uptime.String()))
	if err != nil {
		logger.Error(err.Error())
	}
}

func (a *App) handleAPIServerStatus(w http.ResponseWriter, r *http.Request) {
	dbs := a.getDbStatus()
	appUptime := time.Since(a.startTime).Round(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"dbserver": "%s", "dbversion": "%s", "dbuptime": "%s", "dbstatus": "%s", "appuptime": "%s", "dbtype": "%s"}`, dbs.Name, dbs.Version, dbs.Uptime, dbs.Status, appUptime, dbs.Type)
}

func (a *App) getTagOnDate(tag, date, format string) []byte {
	dateTime, err := a.excelTimeToTime(date)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValue, err := a.store.GetTagDate(tag, dateTime)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}

	switch format {
	case "raw":
		return []byte(fmt.Sprintf("%f", tagValue))
	case "json":
		return []byte(fmt.Sprintf("{ \"Value\" : %f }", tagValue))
	default:
		return []byte(a.store.Format(a.store.Round(tagValue)))
	}
	// if format == "raw" {
	// 	return []byte(fmt.Sprintf("%f", tagValue))
	// } else {
	// 	return []byte(a.store.Format(a.store.Round(tagValue)))
	// }
}

func (a *App) getTagByCount(tag, from, to, count string) []byte {
	fromT, err := a.excelTimeToTime(from)
	if err != nil {
		return []byte(err.Error())
	}
	toT, err := a.excelTimeToTime(to)
	if err != nil {
		return []byte(err.Error())
	}
	countT, err := strconv.Atoi(count)
	if err != nil {
		return []byte(err.Error())
	}
	tagValue, err := a.store.GetTagCount(tag, fromT, toT, countT)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValueJSON, err := json.MarshalIndent(tagValue, "", "  ")
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	return tagValueJSON
}

func (a *App) getTagFromToByCountWithGroup(tag, from, to, count string, group string) []byte {
	fromT, err := a.excelTimeToTime(from)
	if err != nil {
		return []byte(err.Error())
	}
	toT, err := a.excelTimeToTime(to)
	if err != nil {
		return []byte(err.Error())
	}
	countT, err := strconv.Atoi(count)
	if err != nil {
		return []byte(err.Error())
	}
	tagValue, err := a.store.GetTagCountGroup(tag, fromT, toT, countT, group)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValueJSON, err := json.MarshalIndent(tagValue, "", "  ")
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	return tagValueJSON
}

func (a *App) getTagFromTo(tag, from, to string) []byte {
	fromT, err := a.excelTimeToTime(from)
	if err != nil {
		return []byte(err.Error())
	}
	toT, err := a.excelTimeToTime(to)
	if err != nil {
		return []byte(err.Error())
	}
	tagValue, err := a.store.GetTagFromTo(tag, fromT, toT)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValueJSON, err := json.MarshalIndent(tagValue, "", "  ")
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	return tagValueJSON
}

func (a *App) getTagFromToWithGroup(tag, from, to, group string, format string) []byte {
	fromT, err := a.excelTimeToTime(from)

	if err != nil {
		return []byte(err.Error())
	}

	toT, err := a.excelTimeToTime(to)

	if err != nil {
		return []byte(err.Error())
	}
	tagValue, err := a.store.GetTagFromToGroup(tag, fromT, toT, group)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	switch format {
	case "raw":
		return []byte(fmt.Sprintf("%f", tagValue))
	case "json":
		return []byte(fmt.Sprintf("{ \"Value\" : %f }", tagValue))
	default:
		return []byte(a.store.Format(a.store.Round(tagValue)))
	}
}
