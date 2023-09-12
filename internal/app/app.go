package robin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"
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
	round     int
	opCount   int64
	config    config.Config
	cache     cache.BaseCache
	store     store.BaseStore
	template  *template.Template
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
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipnet.IP.IsLoopback() {
				continue
			}
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				localhostIPs = append(localhostIPs, ip4.String())
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
		"/get/tag/":      a.handleGetTag,
		"/get/tag/list/": a.handleGetTagList,
		"/get/tag/up/":   a.handleGetTagUp,
		"/get/tag/down/": a.handleGetTagDown,
		"/favicon.ico":   a.handleFavicon,
		"/api/info/":     a.handleApiInfo,
		"/api/uptime/":   a.handleApiUptime,
		"/api/reload/":   a.handleApiReloadConfig,
		"/api/log/":      a.handleApiGetLog,
		"/api/status/":   a.handleApiServerStatus,
		"/":              a.handleAnyPage("home", nil),
		"/logs/":         a.handleLog,
		"/info/":         a.handleInfo,
		"/health/":       a.handleHealth,
		"/images/":       a.handleDirectory("images"),
		"/scripts/":      a.handleDirectory("scripts"),
		"/css/":          a.handleDirectory("css"),
	}

	// Register HTTP request handlers
	for path, handler := range handlers {
		http.HandleFunc(path, handler)
	}

	// Define custom template function
	funcMap := template.FuncMap{
		"colorizeLogString": func(input string) template.HTML {
			st := strings.Split(input, " ")
			if len(st) > 2 {
				st[0] = "<span class='date'>" + st[0]
				st[1] = st[1] + "</span>"
				st[2] = "<span class='level " + st[2] + "'>" + st[2] + "</span> <span class='level other'>"
			}
			return template.HTML(strings.Join(st, " ") + "</span>")
		},
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

func (a *App) handleApiReloadConfig(w http.ResponseWriter, r *http.Request) {
	logger.Trace("reloading config file")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	logger.Info("reloading config file")

	if err := a.config.Reload(); err != nil {
		logger.Fatal("Failed to read config file")
	}
}

func (a *App) handleApiUptime(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered uptime page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	uptime := time.Since(a.startTime).Round(time.Second)
	_, err := w.Write([]byte(uptime.String()))
	if err != nil {
		logger.Error(err.Error())
	}
}

func (a *App) handleApiServerStatus(w http.ResponseWriter, r *http.Request) {
	version, dbuptimeStr, err := a.store.GetStatus()
	dbstatus := "green"
	if err != nil {
		dbstatus = "red"
	}

	dbname := a.config.GetString("app.db.name")
	dbuptime, err := time.ParseDuration(thenIf(dbuptimeStr != "", dbuptimeStr+"s", "0s"))
	if err != nil {
		logger.Error(err.Error())
	}
	dbuptimeStr = dbuptime.String()

	appUptime := time.Since(a.startTime).Round(time.Second).String()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"dbserver": "%s", "dbversion": "%s", "dbuptime": "%s", "dbstatus": "%s", "appuptime": "%s"}`, dbname, version, dbuptimeStr, dbstatus, appUptime)
}

func (a *App) handleApiInfo(w http.ResponseWriter, r *http.Request) {
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

// handleGetTag handles GET requests for a tag and outputs the corresponding value.
// The tag can be filtered by date, or by a time range. The output can be formatted
// as raw or rounded. The function takes in an http.ResponseWriter and an http.Request
// as parameters, and returns nothing.
func (a *App) handleGetTag(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	headers := w.Header()
	headers.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	headers.Set("Content-Type", "application/json")

	// Increment operation count
	a.opCount++

	// Parse query parameters
	query := r.URL.Query()
	tag := query.Get("tag")
	date := query.Get("date")
	from := query.Get("from")
	to := query.Get("to")
	group := query.Get("group")
	round := query.Get("round")
	count := query.Get("count")
	format := query.Get("format")

	// Set round field of App struct
	a.round = 2
	if round != "" {
		a.round, _ = strconv.Atoi(round)
	}

	// Handle tag and date parameters
	if tag != "" && date != "" {
		dateTime, err := a.excelTimeToTime(date)
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		tagValue, err := a.store.GetTagDate(tag, dateTime)
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		if format == "raw" {
			w.Write([]byte(fmt.Sprintf("%f", tagValue)))
		} else {
			w.Write([]byte(a.store.RoundAndFormat(tagValue)))
		}
		return
	}

	// Handle tag, from, and to parameters
	if tag != "" && from != "" && to != "" {
		fromT, err := a.excelTimeToTime(from)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		toT, err := a.excelTimeToTime(to)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		switch {
		case count != "":
			countT, err := strconv.Atoi(count)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			tagValue, err := a.store.GetTagCount(tag, fromT, toT, countT)
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			tagValueJSON, err := json.MarshalIndent(tagValue, "", "  ")
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			w.Write(tagValueJSON)
			return
		case group == "":
			tagValue, err := a.store.GetTagFromTo(tag, fromT, toT)
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			tagValueJSON, err := json.MarshalIndent(tagValue, "", "  ")
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			w.Write(tagValueJSON)
			return
		default:
			tagValue := a.store.GetTagFromToGroup(tag, fromT, toT, group)
			w.Write([]byte(tagValue))
		}
	}
}

// returns JSON with tags by mask
func (a *App) handleGetTagList(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	// Extract query parameter
	like := r.URL.Query().Get("like")

	// Retrieve list of tags
	tags, err := a.store.GetTagList(like)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	// Marshal tags into JSON string
	j, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		// logger.Debug(err.error())
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	// Write JSON string to response
	w.Write(j)
}

// return time.Time and error
func (a *App) excelTimeToTime(timeStr string) (time.Time, error) {

	if timeStr == "" {
		return time.Time{}, errors.ErrInvalidDate
	}

	var result time.Time

	if !strings.Contains(timeStr, ":") {
		timeStr = strings.Replace(timeStr, ",", ".", 1)
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return time.Time{}, errors.ErrNotAFloat
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
		return time.Time{}, errors.ErrInvalidDate
	}

	return result, nil
}

func (a *App) tryParseDate(date string) (time.Time, error) {
	// if date is empty, return error
	if date == "" {
		return time.Time{}, errors.ErrInvalidDate
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
	return time.Time{}, errors.ErrInvalidDate
}

func (a *App) handleGetTagDown(w http.ResponseWriter, r *http.Request) {
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

func (a *App) handleGetTagUp(w http.ResponseWriter, r *http.Request) {
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

// output program log
func (a *App) handleApiGetLog(w http.ResponseWriter, r *http.Request) {
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
	http.ServeFile(w, r, a.workDir+"/web/images/icon3.png")
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

var logData []string

func (a *App) handleLog(w http.ResponseWriter, r *http.Request) {
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

	a.handleAnyPage(page, getOnePage(page, logData, pageNum, logPerPage))(w, r)

}

func getOnePage(name string, data []string, pageNum, linesPerPage int) map[string]interface{} {
	pagesTotal := len(data) / linesPerPage
	if pageNum > pagesTotal+1 {
		pageNum = pagesTotal + 1
	}

	pageSwicher := generatePageSwitcherHTML(name, pageNum, pagesTotal+1)

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
	if pagerName == "" {
		return fmt.Sprintf(`<button class='page-number %s' href='#' onclick='loadPage("/%s?page=%d")'>%d</button>`,
			thenIf(isCurr, "page-number-current", ""), name, pageNum, pageNum)
	}
	return fmt.Sprintf(`<button class='page-number %s' href='#' onclick='loadPage("/%s?page=%d")'>%s</button>`,
		thenIf(isCurr, "page-number-current", ""), name, pageNum, pagerName)
}

// getDataSubset determines the subset of data to be displayed on the requested page.
func getDataSubset(data []string, pageNum, linesPerPage int) []string {
	startIndex := (pageNum - 1) * linesPerPage
	endIndex := min(pageNum*linesPerPage, len(data))

	return data[startIndex:endIndex]
}

func (a *App) handleInfo(w http.ResponseWriter, r *http.Request) {

	page := "info"

	linesPerPage := 23

	q := r.URL.Query()
	pageNumStr := q.Get("page")
	if pageNumStr == "" {
		tagList = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	tags := map[string]map[time.Time]float32{}
	if q.Get("tag") != "" && q.Get("from") != "" && q.Get("to") != "" {
		from, _ := time.Parse("2006-01-02T15:04", q.Get("from"))
		to, _ := time.Parse("2006-01-02T15:04", q.Get("to"))
		tags, err = a.store.GetTagFromTo(q.Get("tag"), from, to)
		if err != nil {
			fmt.Println("Ошибка при чтении ответа:", err)
			return
		}
	}

	infoData := []string{}
	for _, tag := range tags {
		for d, v := range tag {
			infoData = append(infoData, fmt.Sprintf("%s: %f", d, v))
		}
	}

	// data :=map[string]interface{}{
	// 	page: infoData,
	// }

	a.handleAnyPage(page, getOnePage(page, infoData, pageNum, linesPerPage))(w, r)

}

func (a *App) handleAnyPage(page string, data map[string]interface{}) func(w http.ResponseWriter, r *http.Request) {
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
		// fmt.Printf("%s", t)
		apiserver := "http://" + r.Host
		dataFull := map[string]interface{}{
			"content": t,
			"app":     map[string]interface{}{"name": a.name, "version": a.version, "apiserver": apiserver},
		}

		if err := a.template.ExecuteTemplate(w, "base.html", dataFull); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
		}
	}
}

var tagList []string

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	page := "health"
	linesPerPage := 23

	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		tagList = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if tagList == nil {
		tags, err := a.store.GetTagList("")
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		tagList = append(tagList, tags["tags"]...)
		// for _, tag := range tags["tags"] {
		// 	tagList = append(tagList, fmt.Sprintf("<div>%s</div>", tag))
		// }
	}
	a.handleAnyPage(page, getOnePage(page, tagList, pageNum, linesPerPage))(w, r)
}
