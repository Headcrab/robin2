// fix: check swagger descriptions to all endpoints
// todo: add query temlate system
// todo: sessions
// todo: authenticate
// todo: add tests

package robin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"net/http"
	"strconv"
	"time"

	"robin2/internal/cache"
	"robin2/internal/errors"
	"robin2/internal/store"
	"robin2/pkg/config"
	"robin2/pkg/logger"

	"github.com/joho/godotenv"

	_ "robin2/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// generic thernary operator
func thenIf[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// NewApp creates a new instance of the App struct and returns a pointer to it.
func NewApp() *App {
	return &App{}
}

type App struct {
	name      string
	version   string
	startTime time.Time
	workDir   string
	// round     int
	opCount  int64
	config   config.Config
	cache    cache.BaseCache
	store    store.BaseStore
	template *template.Template
}

func (a *App) Run() {
	a.startTime = time.Now()
	a.init()
	logger.Info(a.name + " " + a.version + " is running")
	err := a.store.Connect(a.cache)
	if err != nil {
		logger.Fatal(err.Error())
	}
	for _, ip := range getLocalhostIpAdresses() {
		logger.Info("listening on http://" + ip + ":" + a.config.GetString("app.port"))
	}
	err = http.ListenAndServe(":"+a.config.GetString("app.port"), nil)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func getLocalhostIpAdresses() []string {
	localhostIPs := []string{"127.0.0.1"}
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Error(err.Error())
		return localhostIPs
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					localhostIPs = append(localhostIPs, ip4.String())
				}
			}
		}
	}
	return localhostIPs
}

func (a *App) init() {
	// Set the working directory of the App object
	a.workDir = getWorkDir()

	// Load environment variables from .env and app.env files
	godotenv.Load(a.workDir+"/.env", a.workDir+"/app.env")

	// Set the name and version fields of the App object
	a.name = os.Getenv("PROJECT_NAME")
	a.version = os.Getenv("PROJECT_VERSION")

	// Configure the logger
	logger.Debug("initializing app")
	a.config = config.GetConfig()
	a.cache = cache.NewFactory().NewCache(a.config.GetString("app.cache.type"))
	a.store = store.NewFactory().NewStore(a.config.GetString("app.db.type"))

	// Define HTTP request handlers
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/get/tag/":      a.handleAPIGetTag,
		"/get/tag/list/": a.handleAPIGetTagList,
		"/get/tag/up/":   a.handleAPIGetTagUp,
		"/get/tag/down/": a.handleAPIGetTagDown,
		"/favicon.ico":   a.handleFavicon,
		"/api/info/":     a.handleAPIInfo,
		"/api/uptime/":   a.handleAPIUptime,
		"/api/reload/":   a.handleAPIReloadConfig,
		"/api/log/":      a.handleAPIGetLog,
		"/api/status/":   a.handleAPIServerStatus,
		"/logs/":         a.handlePageLog,
		"/data/":         a.handlePageData,
		"/tags/":         a.handlePageTags,
		"/":              a.handlePageAny("home", nil),
		"/images/":       a.handleDirectory("images"),
		"/scripts/":      a.handleDirectory("scripts"),
		"/css/":          a.handleDirectory("css"),
		"/swagger/":      httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")),
	}

	// Register HTTP request handlers
	for path, handler := range handlers {
		http.HandleFunc(path, handler)
	}

	// Define custom template function
	funcMap := template.FuncMap{
		"colorizeLogString": colorizeLogString,
		"formatDataString":  formatDataString,
	}

	// Create template object and parse HTML templates
	a.template = template.New("tmpl").Funcs(funcMap)
	var err error
	a.template, err = a.template.ParseGlob(filepath.Join(a.workDir, "web", "templates", "*.html"))
	if err != nil {
		logger.Fatal(err.Error())
		panic(err)
	}
}

func colorizeLogString(input string) template.HTML {
	st := strings.Split(input, " ")
	if len(st) > 2 {
		st[0] = "<span class='date'>" + st[0]
		st[1] = st[1] + "</span>"
		st[2] = "<span class='level " + st[2] + "'>" + st[2] + "</span> <span class='level other'>"
	}
	return template.HTML(strings.Join(st, " ") + "</span>")
}

func formatDataString(input string) template.HTML {
	st := strings.Split(input, "|")
	if len(st) > 1 {
		tm, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", st[0])
		if err != nil {
			logger.Error(err.Error())
		}
		st[0] = tm.Format("01.02.2006 15:04:05")
		st[0] = fmt.Sprintf("<span class='text-center list-group text-list-item level other'>%s</span>", st[0])
		flValue, err := strconv.ParseFloat(st[1], 64)
		if err != nil {
			logger.Error(err.Error())
		}
		st[1] = fmt.Sprintf("<span class='text-center list-group text-list-item level other'>%.2f</span>", flValue)
	}
	return template.HTML(strings.Join(st, " "))
}

func getWorkDir() string {
	executablePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Trace(err.Error())
		return ""
	}

	dir := filepath.Dir(executablePath)
	logger.Trace("working dir set to: " + dir)

	return dir
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

type dbstatus struct {
	Status  string
	Name    string
	Type    string
	Version string
	Uptime  time.Duration
}

func (a *App) getDbStatus() dbstatus {
	dbstatus := dbstatus{
		Status: "green",
		Name:   a.config.GetString("app.db.name"),
		Type:   a.config.GetString("app.db.type"),
	}
	var err error
	var dbuptimeStr string
	dbstatus.Version, dbuptimeStr, err = a.store.GetStatus()
	if err != nil {
		dbstatus.Status = "red"
	}
	dbstatus.Uptime, err = time.ParseDuration(thenIf(dbuptimeStr != "", dbuptimeStr+"s", "0s"))
	if err != nil {
		dbstatus.Status = "red"
	}
	return dbstatus
}

func (a *App) handleAPIServerStatus(w http.ResponseWriter, r *http.Request) {
	dbs := a.getDbStatus()
	appUptime := time.Since(a.startTime).Round(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"dbserver": "%s", "dbversion": "%s", "dbuptime": "%s", "dbstatus": "%s", "appuptime": "%s", "dbtype": "%s"}`, dbs.Name, dbs.Version, dbs.Uptime, dbs.Status, appUptime, dbs.Type)
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

func (a *App) getTagAndDate(tag, date, format string) []byte {
	dateTime, err := a.excelTimeToTime(date)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	tagValue, err := a.store.GetTagDate(tag, dateTime)
	if err != nil {
		return []byte("#Error: " + err.Error())
	}
	if format == "raw" {
		return []byte(fmt.Sprintf("%f", tagValue))
	} else {
		return []byte(a.store.Format(a.store.Round(tagValue)))
	}
}

func (a *App) getTagAndCount(tag, from, to, count string) []byte {
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

func (a *App) getTagAndCountGroup(tag, from, to, count string, group string) []byte {
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

func (a *App) getTagFromToGroup(tag, from, to, group string) []byte {
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
	return []byte(a.store.Format(a.store.Round(tagValue)))
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
		writer = a.getTagAndDate(tag, date, format)
		return
	}

	if tag != "" && from != "" && to != "" {
		if count != "" {
			if group == "" {
				writer = a.getTagAndCount(tag, from, to, count)
				return
			} else {
				writer = a.getTagAndCountGroup(tag, from, to, count, group)
				return
			}
		}

		if group == "" {
			writer = a.getTagFromTo(tag, from, to)
			return
		} else {
			writer = a.getTagFromToGroup(tag, from, to, group)
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

// return time.Time and error
func (a *App) excelTimeToTime(timeStr string) (time.Time, error) {

	if timeStr == "" {
		return time.Time{}, errors.InvalidDate
	}

	var result time.Time

	if !strings.Contains(timeStr, ":") {
		timeStr = strings.Replace(timeStr, ",", ".", 1)
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return time.Time{}, errors.NotAFloat
		}

		unixTime := (timeFloat - 25569) * 86400
		utcTime := time.Unix(int64(unixTime), 0).UTC()
		locTime := utcTime.Local()
		result = locTime
	} else {
		res, err := a.tryParseDate(timeStr)
		if err != nil {
			return time.Time{}, err
		}

		result = res.Local()
	}

	if result.IsZero() {
		return time.Time{}, errors.InvalidDate
	}

	return result, nil
}

func (a *App) tryParseDate(date string) (time.Time, error) {
	// if date is empty, return error
	if date == "" {
		return time.Time{}, errors.InvalidDate
	}
	// if date is not empty, try to parse it to time.Time
	// if date is not valid, return error
	cfg := a.config.GetStringSlice("app.date_formats")
	for fm := range cfg {
		t, err := time.ParseInLocation(cfg[fm], date, time.Local)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.InvalidDate
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

func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, a.workDir+"/web/images/icon.png")
}

func (a *App) handleDirectory(d string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// http.ServeFile(w, r, filepath.Join(a.workDir, d))
		basePath := filepath.Join(a.workDir, "web", d)
		filePath := filepath.Join(basePath, r.URL.Path[len("/"+d+"/"):])
		if !strings.HasPrefix(filePath, basePath) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		logger.Trace(filePath)
		http.ServeFile(w, r, filePath)
	}
}

func getOnePage(name string, data []string, pageNum, linesPerPage int) map[string]interface{} {
	pagesTotal := len(data)/(linesPerPage+1) + 1
	// pagesTotal = thenIf(pagesTotal == 0, 1, pagesTotal)
	if pageNum > pagesTotal {
		pageNum = pagesTotal
	}

	pageSwicher := generatePageSwitcherHTML(name, pageNum, pagesTotal)

	dataL := getDataSubset(data, pageNum, linesPerPage)

	return map[string]interface{}{
		name:   dataL,
		"page": pageSwicher,
	}
}

// generatePageSwitcherHTML generates the HTML for the page switcher component.
func generatePageSwitcherHTML(name string, pageNum, pagesTotal int) template.HTML {
	// pageSwitcher := fmt.Sprintf("<span class='text-center list-group text-list-item'>Страница %d из %d <br>", pageNum, pagesTotal)
	pageSwitcher := "<span class='text-center fixed-bottom2'>"
	// show 10 pages only, current must be in list
	if pagesTotal < 2 {
		return template.HTML(pageSwitcher)
	}

	pageSwitcher += getFormattedPageNumber(name, 1, pageNum == 1, "")

	pageLimit := 11
	if pagesTotal <= pageLimit {
		for i := 2; i < pagesTotal; i++ {
			pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
		}
	} else {
		if pageNum < pageLimit-1 {
			for i := 2; i < pageLimit-1; i++ {
				pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
			}
			pageSwitcher += getFormattedPageNumber(name, pageNum+1, false, "»")
		} else if pageNum > pagesTotal-pageLimit+3 {
			pageSwitcher += getFormattedPageNumber(name, pageNum-1, false, "«")
			for i := pagesTotal - pageLimit + 3; i < pagesTotal; i++ {
				pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
			}
		} else {
			pageSwitcher += getFormattedPageNumber(name, pageNum-1, false, "«")
			for i := pageNum - (pageLimit-4)/2; i <= pageNum+(pageLimit-4)/2; i++ {
				pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
			}
			pageSwitcher += getFormattedPageNumber(name, pageNum+1, false, "»")
		}
	}

	pageSwitcher += getFormattedPageNumber(name, pagesTotal, pageNum == pagesTotal, "")

	pageSwitcher += "</span><br>"
	return template.HTML(pageSwitcher)
}

func getFormattedPageNumber(name string, pageNum int, isCurr bool, pagerName string) string {
	return fmt.Sprintf(`<button class='page-number %s' href='#' onclick='loadPage("/%s?page=%d")'>%s</button>`,
		thenIf(isCurr, "page-number-current", ""), name, pageNum,
		thenIf(pagerName == "", fmt.Sprintf("%d", pageNum), pagerName))
}

// getDataSubset determines the subset of data to be displayed on the requested page.
func getDataSubset(data []string, pageNum, linesPerPage int) []string {
	startIndex := (pageNum - 1) * linesPerPage
	endIndex := min(pageNum*linesPerPage, len(data))

	return data[startIndex:endIndex]
}

func (a *App) handlePageAny(page string, data map[string]interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Trace("rendered " + page + " page")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Content-Type", "text/html")

		contentBuffer := new(bytes.Buffer)
		if err := a.template.ExecuteTemplate(contentBuffer, page+".html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c := contentBuffer.String()
		t := template.HTML(c)
		apiserver := "http://" + r.Host
		dbs := a.getDbStatus()
		appUptime := time.Since(a.startTime).Round(time.Second).String()
		// apiserver := "http://" + a.config.GetString("db."+a.config.GetString("app.db.name")+".host") + ":" + a.config.GetString("app.port")
		dataFull := map[string]interface{}{
			"content": t,
			"app":     map[string]interface{}{"name": a.name, "version": a.version, "apiserver": apiserver, "uptime": appUptime},
			"db":      map[string]interface{}{"server": dbs.Name, "type": dbs.Type, "version": dbs.Version, "uptime": dbs.Uptime, "status": dbs.Status},
		}

		if err := a.template.ExecuteTemplate(w, "base.html", dataFull); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
		}
	}
}

var logData []string

func (a *App) handlePageLog(w http.ResponseWriter, r *http.Request) {
	procTimeBegin := time.Now()
	page := "logs"
	logPerPage := 23
	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		pageNumStr = "1"
		logData = nil
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if logData == nil {
		logData, err = logger.GetLogHistory()
		if err != nil {
			errorMsg := "#Error: " + err.Error()
			w.Write([]byte(errorMsg))
			logger.Error(errorMsg)
			return
		}
	}
	w.Header().Set("Procession-Time", time.Since(procTimeBegin).String())
	a.handlePageAny(page, getOnePage(page, logData, pageNum, logPerPage))(w, r)

}

var tagsValues map[string]map[time.Time]float32

func (a *App) handlePageData(w http.ResponseWriter, r *http.Request) {
	procTimeBegin := time.Now()
	page := "data"

	linesPerPage := 23

	q := r.URL.Query()
	pageNumStr := q.Get("page")
	if pageNumStr == "" {
		tagsValues = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if tagsValues == nil {
		if q.Get("tag") != "" && q.Get("from") != "" && q.Get("to") != "" {
			from, _ := time.Parse("2006-01-02T15:04", q.Get("from"))
			to, _ := time.Parse("2006-01-02T15:04", q.Get("to"))
			countStr, _ := strconv.Atoi(q.Get("count"))
			count := int(countStr)
			// tags, err = a.store.GetTagFromTo(q.Get("tag"), from, to)
			tagsValues, err = a.store.GetTagCount(q.Get("tag"), from, to, count)
			if err != nil {
				fmt.Println("Ошибка при чтении ответа:", err)
				return
			}
		}
	}

	data := []string{}
	for _, tag := range tagsValues {
		times := make([]time.Time, 0, len(tag))
		for k := range tag {
			times = append(times, k)
		}
		sort.Slice(times, func(i, j int) bool {
			return times[i].Before(times[j])
		})
		for _, v := range times {
			// data = append(data, fmt.Sprintf("%s: %f", v, tag[v]))
			data = append(data, fmt.Sprintf("%s|%f", v, tag[v]))
		}
	}

	w.Header().Set("Procession-Time", time.Since(procTimeBegin).String())
	a.handlePageAny(page, getOnePage(page, data, pageNum, linesPerPage))(w, r)

}

var tagsList []string

func (a *App) handlePageTags(w http.ResponseWriter, r *http.Request) {
	procTimeBegin := time.Now()
	page := "tags"
	linesPerPage := 23

	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		tagsList = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	like := r.URL.Query().Get("like")
	if like != "" {
		if tagsList == nil {
			tags, err := a.store.GetTagList(like)
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			tagsList = append(tagsList, tags["tags"]...)
		}
	}

	data := getOnePage(page, tagsList, pageNum, linesPerPage)
	data["like"] = thenIf(like == "", "", like)

	w.Header().Set("Procession-Time", time.Since(procTimeBegin).String())
	a.handlePageAny(page, data)(w, r)
}
