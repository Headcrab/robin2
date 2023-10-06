// fix: check swagger descriptions to all endpoints
// todo: add query temlate system
// todo: sessions
// todo: authenticate
// todo: add tests

package robin

import (
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

	_ "robin2/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	// round     int
	opCount  int64
	config   config.Config
	cache    cache.BaseCache
	store    store.BaseStore
	template *template.Template
}

func (a *App) Run() {
	a.startTime = time.Now()
	a.workDir = getWorkDir()
	godotenv.Load(a.workDir+"/.env", a.workDir+"/app.env")
	a.name = os.Getenv("PROJECT_NAME")
	a.version = os.Getenv("PROJECT_VERSION")
	logger.Info(a.name + " " + a.version + " is running")

	a.initApp()
	err := a.store.Connect(a.cache)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("listening on: " + strings.Join(getLocalhostIpAdresses(), " , ") + " port: " + a.config.GetString("app.port"))
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

func (a *App) initApp() {
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
		"/api/info/":     a.handleAPIInfo,
		"/api/uptime/":   a.handleAPIUptime,
		"/api/reload/":   a.handleAPIReloadConfig,
		"/api/log/":      a.handleAPIGetLog,
		"/api/status/":   a.handleAPIServerStatus,
		"/favicon.ico":   a.handleFavicon,
		"/logs/":         a.handlePageLog,
		"/data/":         a.handlePageData,
		"/tags/":         a.handlePageTags,
		"/":              a.handlePageAny("home", nil),
		"/images/":       a.handleDirectory("images"),
		"/scripts/":      a.handleDirectory("scripts"),
		"/css/":          a.handleDirectory("css"),
		"/swagger/":      httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")),
		"/templ/list/":   a.handleTemplateList,
		"/templ/add/":    a.handleTemplateAdd,
		"/templ/get/":    a.handleTemplateGet,
		"/templ/edit/":   a.handleTemplateEdit,
		"/templ/delete/": a.handleTemplateDelete,
		"/templ/exec/":   a.handleTemplateExec,
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
		st[0] = fmt.Sprintf("<td> <span class='text-center list-group text-list-item level other'>%s</span> </td>", st[0])
		flValue, err := strconv.ParseFloat(st[1], 64)
		if err != nil {
			logger.Error(err.Error())
		}
		st[1] = fmt.Sprintf("<td> <span class='text-center list-group text-list-item level other'>%.2f</span> </td>", flValue)
	}
	return template.HTML("<tr>" + strings.Join(st, " ") + "</tr>")
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

// generic thernary operator
func thenIf[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}
