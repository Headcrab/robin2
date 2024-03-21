package robin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"robin2/internal/data"
	"robin2/internal/decode"
	"robin2/internal/format"
	"robin2/internal/logger"
	"robin2/internal/utils"
	"strconv"
	"strings"
	"time"
)

// @Summary Получить лог
// @Description Возвращает логи приложения
// @Tags System
// @Produce json
// @Success 200 {array} string
// @Router /api/log/ [get]
// @Param format query string false "Формат вывода (str - по умолчанию, json, raw)"
func (a *App) handleAPIGetLog(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered log page")
	// w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	// w.Header().Set("Content-Type", "application/json")
	formatStr := r.URL.Query().Get("format")
	logs, err := logger.GetLogHistory()
	if err != nil {
		if _, err := w.Write([]byte("#Error: " + err.Error())); err != nil {
			logger.Error(fmt.Sprintf("Ошибка при записи ответа: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	tagValue := format.New(formatStr).Process(logs)
	// tagValue, err := json.MarshalIndent(logs, "", "  ")
	// if err != nil {
	// 	w.Write([]byte("#Error: " + err.Error()))
	// 	return
	// }

	if _, err := w.Write(tagValue); err != nil {
		logger.Error(fmt.Sprintf("Ошибка при записи ответа: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
// @Param format query string false "Формат вывода (str - по умолчанию, json, raw)"
func (a *App) handleAPIGetTag(w http.ResponseWriter, r *http.Request) {
	writer := []byte("Error: unknown error")
	defer func() {
		if _, err := w.Write(writer); err != nil {
			logger.Error(fmt.Sprintf("Ошибка при записи ответа: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	a.opCount++

	query := r.URL.Query()
	tag := query.Get("tag")
	date := query.Get("date")
	from := query.Get("from")
	to := query.Get("to")
	group := query.Get("group")
	roundStr := query.Get("round")
	count := query.Get("count")
	format := query.Get("format")

	round := utils.ThenIf(roundStr != "", func() int { r, _ := strconv.Atoi(roundStr); return r }(), a.config.Round)

	if tag != "" && date != "" {
		tags := strings.Split(tag, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		if len(tags) > 1 {
			writer = a.getTagsOnDate(tags, date, format, round)
			return
		} else {
			writer = a.getTagOnDate(tag, date, format, round)
			return
		}
	}

	if tag != "" && from != "" && to != "" {
		if count != "" {
			if group == "" {
				writer = a.getTagByCount(tag, from, to, count, format, round)
				return
			} else {
				writer = a.getTagFromToByCountWithGroup(tag, from, to, count, group, format, round)
				return
			}
		}

		if group == "" {
			writer = a.getTagFromTo(tag, from, to, format, round)
			return
		} else {
			writer = a.getTagFromToWithGroup(tag, from, to, group, format, round)
			return
		}
	}
}

// @Summary Получить список тегов
// @Description Возвращает список всех тегов по маске
// @Tags Tag
// @Produce plain/text
// @Success 200 {array} string
// @Router /get/tag/list/ [get]
// @Param like query string false "Маска поиска"
// @Param format query string false "Формат вывода (str - по умолчанию, json, raw)"
func (a *App) handleAPIGetTagList(w http.ResponseWriter, r *http.Request) {
	writer := []byte("#Error: unknown error")
	defer func() {
		if _, err := w.Write(writer); err != nil {
			logger.Error(fmt.Sprintf("Ошибка при записи ответа: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	// Extract query parameter
	like := r.URL.Query().Get("like")
	fmt := r.URL.Query().Get("format")
	roundStr := r.URL.Query().Get("round")
	round := utils.ThenIf(roundStr != "", func() int { r, _ := strconv.Atoi(roundStr); return r }(), a.config.Round)

	// Retrieve list of tags
	tags, err := a.store.GetTagList(like)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	writer = format.New(fmt).SetRound(round).Process(tags)
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
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		if _, err := w.Write([]byte("#Error: tag is empty")); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 0
	}
	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)
	if err != nil {
		if _, err = w.Write([]byte("#Error: " + err.Error())); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)
	if err != nil {
		if _, err = w.Write([]byte("#Error: " + err.Error())); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	v, err := a.store.GetDownDates(tag, fromT, toT)
	if err != nil {
		if _, err = w.Write([]byte("#Error: " + err.Error())); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	} else {
		// if err != nil {
		// 	if _,err = w.Write([]byte("#Error: " + err.Error())); err != nil {
		// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// 	}
		// 	return
		// }
		if (count >= 0) && (count < len(v)) {
			val := v[count].Format("2006-01-02 15:04:05")
			if _, err = w.Write([]byte(val)); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			if _, err = w.Write([]byte("")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
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
	writer := []byte("#Error: unknown error")
	defer func() {
		if _, err := w.Write(writer); err != nil {
			logger.Error(fmt.Sprintf("Ошибка при записи ответа: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	tag := r.URL.Query().Get("tag")
	if tag == "" {
		writer = []byte("#Error: tag is empty")
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 0
	}

	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	v, err := a.store.GetUpDates(tag, fromT, toT)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	if count >= 0 && count < len(v) {
		val := v[count].Format("2006-01-02 15:04:05")
		writer = []byte(val)
		return
	} else {
		writer = []byte("")
		return
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

// @Summary Презагрузить конфигурационный файл
// @Description Презагружает конфигурационный файл приложения в случае изменения
// @Tags System
// @Produce plain/text
// @Success 200 {array} string
// @Router /api/reload/ [get]
func (a *App) handleAPIReloadConfig(w http.ResponseWriter, r *http.Request) {
	logger.Trace("reloading config file")

	if err := a.config.Reload(); err != nil {
		logger.Fatal("Failed to read config file")
	}
	a.initDatabase()
}

func (a *App) handleAPIServerStatus(w http.ResponseWriter, r *http.Request) {
	dbs := a.getDbStatus()
	appUptime := time.Since(a.startTime).Round(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"dbserver": "%s", "dbversion": "%s", "dbuptime": "%s", "dbstatus": "%s", "appuptime": "%s", "dbtype": "%s"}`, dbs.Name, dbs.Version, dbs.Uptime, dbs.Status, appUptime, dbs.Type)
}

func (a *App) getTagOnDate(tag, date, fmt string, round int) []byte {
	dateTime, err := utils.ExcelTimeToTime(date, a.config.DateFormats)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValue, err := a.store.GetTagDate(tag, dateTime)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}

	w := format.New(fmt).SetRound(round).Process(tagValue)
	return w
}

func (a *App) getTagsOnDate(tags []string, date, fmt string, round int) []byte {
	dateTime, err := utils.ExcelTimeToTime(date, a.config.DateFormats)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}

	tagsVal := data.Tags{}
	for _, tag := range tags {
		tagValue, err := a.store.GetTagDate(tag, dateTime)
		if err != nil {
			continue
		}
		tagsVal = append(tagsVal, tagValue)
	}
	w := format.New(fmt).SetRound(round).Process(tagsVal)
	return w
}

func (a *App) getTagByCount(tag, from, to, count string, fmt string, round int) []byte {
	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)
	if err != nil {
		return []byte(err.Error())
	}
	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)
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

	w := format.New(fmt).SetRound(round).Process(tagValue)
	return w
}

func (a *App) getTagFromToByCountWithGroup(tag, from, to, count string, group string, fmt string, round int) []byte {

	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)
	if err != nil {
		return []byte(err.Error())
	}

	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)
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

	w := format.New(fmt).SetRound(round).Process(tagValue)
	return w
}

func (a *App) getTagFromTo(tag, from, to string, fmt string, round int) []byte {
	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)
	if err != nil {
		return []byte(err.Error())
	}
	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)
	if err != nil {
		return []byte(err.Error())
	}
	tagValue, err := a.store.GetTagFromTo(tag, fromT, toT)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	w := format.New(fmt).SetRound(round).Process(tagValue)
	return w
}

func (a *App) getTagFromToWithGroup(tag, from, to, group string, fmt string, round int) []byte {
	fromT, err := utils.ExcelTimeToTime(from, a.config.DateFormats)

	if err != nil {
		return []byte(err.Error())
	}

	toT, err := utils.ExcelTimeToTime(to, a.config.DateFormats)

	if err != nil {
		return []byte(err.Error())
	}
	tags := strings.Split(tag, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	tdv := make(map[string]map[time.Time]float32)
	for _, tag := range tags {
		tdv[tag] = make(map[time.Time]float32)
		tdv[tag][toT], err = a.store.GetTagFromToGroup(tag, fromT, toT, group)
		if err != nil {
			return []byte("#Error: " + err.Error())
		}

	}
	var w []byte
	if len(tdv) == 1 {
		w = format.New(fmt).SetRound(round).Process(tdv[tags[0]][toT])
	} else {
		w = format.New(fmt).SetRound(round).Process(tdv)
	}
	return w
}

// @Summary Получить расшифровку имени тега
// @Description Возвращает расшифровку имени тега
// @Tags Tag
// @Produce plain/text json
// @Success 200 {array} string
// @Router /tag/decode/ [get]
// @Param tag query string true "Наименование тега"
// @Param format query string false "Формат вывода (str - по умолчанию, json, raw)"
func (a *App) handleTagDecode(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	formatStr := r.URL.Query().Get("format")

	if tag == "" {
		if _, err := w.Write([]byte("#Error: tag is empty")); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	tags := strings.Split(tag, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}

	ret := make(map[string]map[string]string)
	var dec decode.Decoder
	err := dec.LoadJSONData(filepath.Join(a.workDir, "config"))
	if err != nil {
		if _, err := w.Write([]byte("#Error: " + err.Error())); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	for _, t := range tags {
		dec.Tags = append(dec.Tags, decode.Tag{Name: t})
		dec.DecodeTags()
	}
	for item := range dec.DecodedTagsChan {
		ret[item["tag_name"]] = item
	}

	s := format.New(formatStr).Process(ret)
	if _, err := w.Write([]byte(s)); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
