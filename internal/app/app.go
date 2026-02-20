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
	"sync"
	"time"

	"robin2/internal/cache"
	"robin2/internal/config"
	"robin2/internal/data"
	"robin2/internal/format"
	"robin2/internal/logger"
	"robin2/internal/middleware"
	"robin2/internal/store"
	"robin2/internal/utils"

	"github.com/joho/godotenv"

	_ "robin2/docs"
)

type App struct {
	name          string
	version       string
	startTime     time.Time
	workDir       string
	opCount       int64
	config        config.Config
	cache         cache.Cache
	store         store.Store
	template      *template.Template
	formatterPool *format.FormatterPool
	pageCache     pageCache
	dbStatusCache dbStatusCache
}

type dbStatus struct {
	Status  string
	Name    string
	Type    string
	Version string
	Uptime  time.Duration
}

type pageCache struct {
	mu         sync.Mutex
	logData    []string
	tagsValues data.Tags
	tagsList   []string
}

type dbStatusCache struct {
	mu        sync.RWMutex
	ttl       time.Duration
	value     dbStatus
	expiresAt time.Time
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
	app.formatterPool = format.NewFormatterPool(10)
	app.dbStatusCache.ttl = 5 * time.Second
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
	a.invalidateDbStatusCache()

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
		"/get/tag/":             a.handleAPIGetTag,
		"/get/tag/list/":        a.handleAPIGetTagList,
		"/get/tag/up/":          a.handleAPIGetTagUp,
		"/get/tag/down/":        a.handleAPIGetTagDown,
		"/api/info/":            a.handleAPIInfo,
		"/api/reload/":          a.handleAPIReloadConfig,
		"/api/log/":             a.handleAPIGetLog,
		"/api/log/clear/":       a.handleAPIClearLog,
		"/api/status/":          a.handleAPIServerStatus,
		"/favicon.ico":          a.handleFavicon,
		"/logs/":                a.handlePageLog,
		"/data/":                a.handlePageData,
		"/tags/":                a.handlePageTags,
		"/docs/":                a.handlePageDocs,
		"/docs/view/":           a.handleDocView,
		"/":                     a.handlePageAny("home", map[string]interface{}{"descr": "Robin"}),
		"/images/":              a.handleDirectory("images"),
		"/scripts/":             a.handleDirectory("scripts"),
		"/css/":                 a.handleDirectory("css"),
		"/api/swagger/":         a.handleSwaggerDark,
		"/api/swagger/doc.json": a.handleSwaggerJSON,
		"/swagger/":             a.handlePageSwagger,
		"/templ/list/":          a.handleTemplateList,
		"/templ/add/":           a.handleTemplateAdd,
		"/templ/get/":           a.handleTemplateGet,
		"/templ/edit/":          a.handleTemplateEdit,
		"/templ/delete/":        a.handleTemplateDelete,
		"/templ/exec/":          a.handleTemplateExec,
		"/tag/decode/":          a.handleTagDecode,
		"/api/v2/get/":          a.handleAPIV2GetTagOnDate,
	}

	// Register HTTP request handlers
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}

	// Define custom template function
	funcMap := template.FuncMap{
		"colorizeLogString": colorizeLogString,
		"formatDataString":  formatDataString,
		"div":               func(a, b float64) float64 { return a / b },
		"formatFileSize":    func(size int64) string { return fmt.Sprintf("%.1f KB", float64(size)/1024.0) },
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
	if cached, ok := a.readDBStatusCache(); ok {
		return cached
	}

	dbName := a.config.CurrDB.Name
	dbstatus := dbStatus{
		Status: "green",
		Name:   dbName,
		Type:   a.config.CurrDB.Type,
	}

	// проверяем что store инициализирован
	if a.store == nil {
		dbstatus.Status = "red"
		dbstatus.Version = "unknown"
		dbstatus.Uptime = 0
		return dbstatus
	}

	var err error
	dbstatus.Version, dbstatus.Uptime, err = a.store.GetStatus()
	if err != nil {
		dbstatus.Status = "red"
	}

	a.writeDBStatusCache(dbstatus)
	return dbstatus
}

func (a *App) readDBStatusCache() (dbStatus, bool) {
	a.dbStatusCache.mu.RLock()
	defer a.dbStatusCache.mu.RUnlock()

	if a.dbStatusCache.expiresAt.IsZero() || time.Now().After(a.dbStatusCache.expiresAt) {
		return dbStatus{}, false
	}
	return a.dbStatusCache.value, true
}

func (a *App) writeDBStatusCache(status dbStatus) {
	a.dbStatusCache.mu.Lock()
	a.dbStatusCache.value = status
	a.dbStatusCache.expiresAt = time.Now().Add(a.dbStatusCache.ttl)
	a.dbStatusCache.mu.Unlock()
}

func (a *App) invalidateDbStatusCache() {
	a.dbStatusCache.mu.Lock()
	a.dbStatusCache.value = dbStatus{}
	a.dbStatusCache.expiresAt = time.Time{}
	a.dbStatusCache.mu.Unlock()
}

// handleSwaggerDark serves Swagger UI with theme detection
func (a *App) handleSwaggerDark(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        /* Hide top bar always */
        .topbar {
            display: none !important;
        }
        
        /* Base light theme styles */
        body {
            margin: 0;
            padding: 0;
        }
        
        /* Dark theme styles */
        body.dark-theme {
            background-color: #1a1a1a !important;
            color: #e1e1e1 !important;
        }
        
        body.dark-theme .swagger-ui {
            filter: invert(1) hue-rotate(180deg);
        }
        
        body.dark-theme .swagger-ui img,
        body.dark-theme .swagger-ui .swagger-ui img {
            filter: invert(1) hue-rotate(180deg);
        }
        
        body.dark-theme .swagger-ui .scheme-container {
            background: #2d2d2d !important;
        }
        
        body.dark-theme .swagger-ui .info {
            margin-bottom: 20px;
        }
        
        /* Light theme styles */
        body.light-theme {
            background-color: #ffffff !important;
            color: #333333 !important;
        }
        
        body.light-theme .swagger-ui {
            filter: none;
        }
        
        body.light-theme .swagger-ui .scheme-container {
            background: #fafafa !important;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        // Detect theme from parent window or localStorage
        function detectTheme() {
            try {
                // Try to get theme from parent window (if in iframe)
                if (window.parent && window.parent !== window) {
                    const parentTheme = window.parent.document.documentElement.getAttribute('data-theme');
                    if (parentTheme) {
                        return parentTheme;
                    }
                }
                
                // Try localStorage
                const savedTheme = localStorage.getItem('theme');
                if (savedTheme) {
                    return savedTheme;
                }
                
                // Check system preference
                if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
                    return 'dark';
                }
            } catch (e) {
                console.log('Theme detection failed:', e);
            }
            
            return 'light'; // default
        }
        
        function applyTheme(theme) {
            document.body.className = theme + '-theme';
            console.log('Applied Swagger theme:', theme);
        }
        
        // Apply initial theme
        const currentTheme = detectTheme();
        applyTheme(currentTheme);
        
        // Listen for theme changes from parent window
        window.addEventListener('message', function(event) {
            if (event.data && event.data.type === 'themeChanged') {
                applyTheme(event.data.theme);
            }
        });
        
        // Initialize Swagger UI
        window.onload = function() {
            SwaggerUIBundle({
                url: '/api/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(swaggerHTML))
}

// handleSwaggerJSON serves swagger JSON spec
func (a *App) handleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Read swagger.json from docs folder
	swaggerFile := filepath.Join(a.workDir, "docs", "swagger.json")
	http.ServeFile(w, r, swaggerFile)
}
