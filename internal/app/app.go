package robin

// todo: authenticate (in web only?)
// todo: add tests

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
	cache     cache.Cache
	store     store.Store
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
	err := godotenv.Load(filepath.Join(app.workDir, ".env"))
	if err != nil {
		logger.Info(err.Error())
	}
	app.name = os.Getenv("PROJECT_NAME")
	app.version = os.Getenv("PROJECT_VERSION")
	// app.config = config.New()
	app.config.Load(filepath.Join(app.workDir, "config", "Robin.json"))
	return &app
}

func (a *App) Run() {
	a.startTime = time.Now()
	logger.Info(a.name + " " + a.version + " is running")

	if err := a.initDatabase(); err != nil {
		logger.Fatal(err.Error())
	}

	mux := a.initHTTPHandlers()
	if mux == nil {
		logger.Fatal("Failed to initialize HTTP handlers")
	}

	logger.Info("listening on: " + strings.Join(utils.GetLocalhostIpAdresses(),
		":"+strconv.Itoa(a.config.Port)+", ") + ":" + strconv.Itoa(a.config.Port))

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(a.config.Port),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err.Error())
		}
	}()

	// Add a mechanism to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(err.Error())
	}
}

func (a *App) initDatabase() error {
	var err error
	a.cache, err = cache.New(a.config)
	if err != nil {
		return err
	}
	a.store, err = store.New(a.config)
	if err != nil {
		return err
	}

	err = a.store.Connect("default", a.cache)
	if err != nil {
		return err
	}

	return nil
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
	parts := strings.Split(input, " ")
	if len(parts) > 2 {
		parts[0] = "<span class='date'>" + parts[0]
		parts[1] = parts[1] + "</span>"
		parts[2] = "<span class='level " + parts[2] + "'>" + parts[2] + "</span> <span class='level other'>"
	}
	return template.HTML(strings.Join(parts, " ") + "</span>")
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
	dbstatus.Version, dbstatus.Uptime, err = a.store.GetStatus()
	if err != nil {
		dbstatus.Status = "red"
	}

	return dbstatus
}
