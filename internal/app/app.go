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

	w.Write([]byte("ok"))
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

// func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
// 	a.handleAnyPage("home.html", map[string]interface{}{})(w, r)

// }

func (a *App) handleLog(w http.ResponseWriter, r *http.Request) {
	page := "logs"

	logData, err := logger.GetLogHistory()
	if err != nil {
		errorMsg := "#Error: " + err.Error()
		w.Write([]byte(errorMsg))
		logger.Error(errorMsg)
		return
	}

	logPerPage := 23
	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	a.handleAnyPage(page, getOnePage(page, logData, pageNum, logPerPage))(w, r)

}

func getOnePage(name string, data []string, pageNum, linesPerPage int) map[string]interface{} {
	pagesTotal := len(data) / linesPerPage
	if pageNum > pagesTotal+1 {
		pageNum = pagesTotal + 1
	}

	var pageSwicher template.HTML
	pageSwicher += template.HTML("<span class='text-center list-group text-list-item'>Страница " + strconv.Itoa(pageNum) + " из " + strconv.Itoa(pagesTotal+1) + "<br>")
	for i := 1; i <= pagesTotal+1; i++ {
		pageSwicher += template.HTML("<span class='level other' style='background-color: rgba(0, 123, 255, 0.10); cursor=pointer;'><a href='#' onclick='loadPage(\"/" + name + "?page=" + strconv.Itoa(i) + "\")'>" + strconv.Itoa(i) + "</a></span> ")
		// pageSwicher += template.HTML("<a href='/" + name + "?page=" + strconv.Itoa(i) + "'>" + strconv.Itoa(i) + "</a> ")
	}
	pageSwicher += "</span><br>"

	data = data[(pageNum-1)*linesPerPage : min(pageNum*linesPerPage, len(data))]

	return map[string]interface{}{
		name:   data,
		"page": pageSwicher,
	}
}

func (a *App) handleInfo(w http.ResponseWriter, r *http.Request) {

	page := "info"

	data := map[string]interface{}{
		"Uptime": time.Since(a.startTime).Round(time.Second).String(),
	}

	a.handleAnyPage(page, data)(w, r)

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

		data = map[string]interface{}{
			"content": template.HTML(contentBuffer.String()),
			"App":     map[string]interface{}{"Name": a.name, "Version": a.version},
		}

		if err := a.template.ExecuteTemplate(w, "base.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
		}
	}
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	page := "health"

	var data []string
	for i := 0; i < 100000; i++ {
		data = append(data, "Test line "+strconv.Itoa(i)+" --------- ")
	}

	linesPerPage := 23
	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	a.handleAnyPage(page, getOnePage(page, data, pageNum, linesPerPage))(w, r)

	w.Write([]byte("OK"))
}
