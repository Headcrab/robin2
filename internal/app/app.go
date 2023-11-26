// todo: authenticate (in web only?)
// todo: add tests

package robin

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"net/http"
	"strconv"
	"time"

	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/logger"
	"robin2/internal/middleware"
	"robin2/internal/store"
	"robin2/internal/utils"

	"github.com/joho/godotenv"
	swagger "github.com/swaggo/http-swagger/v2"

	_ "robin2/docs"
)

type App struct {
	name      string
	version   string
	startTime time.Time
	workDir   string
	opCount   int64
	config    config.Config
	cache     cache.BaseCache
	store     store.BaseStore
	template  *template.Template
}

type dbStatus struct {
	Status  string
	Name    string
	Type    string
	Version string
	Uptime  time.Duration
}

func NewApp() *App {
	app := App{}
	logger.Debug("initializing app")
	app.workDir = utils.GetWorkDir()
	godotenv.Load(filepath.Join(app.workDir, ".env"), filepath.Join(app.workDir, "app.env"))
	app.name = os.Getenv("PROJECT_NAME")
	app.version = os.Getenv("PROJECT_VERSION")
	app.config.Load(filepath.Join(app.workDir, "config", "robin.json"))
	return &app
}

func (a *App) Run() {
	a.startTime = time.Now()
	logger.Info(a.name + " " + a.version + " is running")

	a.initDatabase()

	mux := a.initHTTPHandlers()
	logger.Info("listening on: " + strings.Join(utils.GetLocalhostIpAdresses(), " , ") + " port: " + strconv.Itoa(a.config.Port))

	err := http.ListenAndServe(":"+strconv.Itoa(a.config.Port), mux)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (a *App) initDatabase() {
	a.cache = cache.NewFactory().NewCache(a.config.CurrCache.Type, a.config)
	a.store = store.NewFactory().NewStore(a.config.CurrDB.Type, a.config)

	err := a.store.Connect(a.config.CurrDB.Name, a.cache)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (a *App) initHTTPHandlers() http.Handler {
	// a.template = template.New("tmpl")
	mux := http.NewServeMux()
	// Define HTTP request handlers
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/get/tag/":      a.handleAPIGetTag,
		"/get/tag/list/": a.handleAPIGetTagList,
		"/get/tag/up/":   a.handleAPIGetTagUp,
		"/get/tag/down/": a.handleAPIGetTagDown,
		"/api/info/":     a.handleAPIInfo,
		"/api/reload/":   a.handleAPIReloadConfig,
		"/api/log/":      a.handleAPIGetLog,
		"/api/status/":   a.handleAPIServerStatus,
		"/favicon.ico":   a.handleFavicon,
		"/logs/":         a.handlePageLog,
		"/data/":         a.handlePageData,
		"/tags/":         a.handlePageTags,
		"/":              a.handlePageAny("home", map[string]interface{}{"descr": "Robin"}),
		"/images/":       a.handleDirectory("images"),
		"/scripts/":      a.handleDirectory("scripts"),
		"/css/":          a.handleDirectory("css"),
		"/api/swagger/":  swagger.Handler(swagger.URL("/api/swagger/doc.json")),
		"/swagger/":      a.handlePageSwagger,
		"/templ/list/":   a.handleTemplateList,
		"/templ/add/":    a.handleTemplateAdd,
		"/templ/get/":    a.handleTemplateGet,
		"/templ/edit/":   a.handleTemplateEdit,
		"/templ/delete/": a.handleTemplateDelete,
		"/templ/exec/":   a.handleTemplateExec,
		"/tag/decode/":   a.handleTagDecode,
		"/api/v2/get/":   a.handleAPIV2GetTagOnDate,
	}

	// Register HTTP request handlers
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}

	// timedMux := TimingMiddleware(mux)
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

	return middleware.Log(middleware.Timing(mux))
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

// formatDataString форматирует входную строку в HTML-шаблон.
//
// Функция принимает параметр input типа string, который представляет входные данные для форматирования.
//
// Она возвращает тип template.HTML, который представляет отформатированную строку, преобразованную в HTML-строку таблицы.
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

// getDbStatus возвращает статус базы данных.
//
// Он извлекает имя текущей базы данных из конфигурации приложения.
// Затем он создает структуру dbstatus с именем, типом и статусом "green" по умолчанию.
// Далее он вызывает метод GetStatus из хранилища, чтобы получить версию и время работы базы данных.
// Если происходит ошибка, статус устанавливается на "red".
// Наконец, он преобразует строку времени работы в значение типа duration и устанавливает его в структуре dbstatus.
//
// Возвращает структуру dbstatus, содержащую статус, имя, тип, версию и время работы базы данных.
func (a *App) getDbStatus() dbStatus {
	dbName := a.config.CurrDB.Name
	dbstatus := dbStatus{
		Status: "green",
		Name:   dbName,
		Type:   a.config.CurrDB.Type,
	}
	var err error
	var dbuptimeStr string
	dbstatus.Version, dbuptimeStr, err = a.store.GetStatus()
	if err != nil {
		dbstatus.Status = "red"
	}
	dbstatus.Uptime, err = time.ParseDuration(utils.ThenIf(dbuptimeStr != "", dbuptimeStr+"s", "0s"))
	if err != nil {
		dbstatus.Status = "red"
	}
	return dbstatus
}
