package robin

import (
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

func NewApp() *App {
	app := &App{}
	return app
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

	a.workDir = getWorkDir()
	godotenv.Load(a.workDir+"/.env", a.workDir+"/app.env")
	a.name = os.Getenv("PROJECT_NAME")
	a.version = os.Getenv("PROJECT_VERSION")

	logger.Debug("initializing app")
	a.config = config.GetConfig()
	a.cache = cache.NewFactory().NewCache(a.config.GetString("app.cache.type"))
	a.store = store.NewFactory().NewStore(a.config.GetString("app.db.type"))
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/":               a.handleHome,
		"/info/":          a.handleInfo,
		"/uptime/":        a.handleUptime,
		"/reload_config/": a.handleReloadConfig,
		// "/get/example/":   a.handleGetExample,
		"/get/tag/":      a.handleGetTag,
		"/get/tag/list/": a.handleGetTagList,
		"/get/tag/up/":   a.handleGetTagUp,
		"/get/tag/down/": a.handleGetTagDown,
		"/favicon.ico":   a.handleFavicon,
		"/log/":          a.handleGetLog,
		"/images/":       a.handleImages,
		"/scripts/":      a.handleScripts,
	}
	for path, handler := range handlers {
		http.HandleFunc(path, handler)
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

func (a *App) handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	logger.Info("reloading config file")
	err := a.config.Reload()
	if err != nil {
		logger.Fatal("Failed to read config file")
	}
	w.Write([]byte("ok"))

}

func (a *App) handleUptime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	uptime := time.Since(a.startTime).Round(time.Second)
	_, err := w.Write([]byte(uptime.String()))
	if err != nil {
		logger.Error(err.Error())
	}
}

func (a *App) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	inf := make(map[string]interface{})
	inf["name"] = a.name
	inf["version"] = a.version
	inf["uptime"] = time.Since(a.startTime).Round(time.Second).String()
	inf["op_count"] = a.opCount
	// todo: add more logger.Info
	// inf["config_file_path"] = a.config.ConfigFileUsed()
	// inf["config_file"] = a.config.ConfigFileUsed()
	// inf["config"] = a.config.AllSettings() //todo hide passwords
	resJson, _ := json.MarshalIndent(inf, "", "  ")
	w.Write(resJson)
}

// func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
// 	f, _ := os.ReadFile("index.html")
// 	w.Write(f)
// }

// handleGetTag handles GET requests for a tag and outputs the corresponding
// value. The tag can be filtered by date, or by a time range. The output can
// be formatted as raw or rounded. The function takes in an http.ResponseWriter
// and an http.Request as parameters, and returns nothing.
func (a *App) handleGetTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	a.opCount++
	tag := r.URL.Query().Get("tag")
	date := r.URL.Query().Get("date")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	group := r.URL.Query().Get("group")
	round := r.URL.Query().Get("round")
	count := r.URL.Query().Get("count")
	format := r.URL.Query().Get("format")
	// fmt.Println("tag:", r.URL.Query()["tag"])
	if round != "" {
		a.round, _ = strconv.Atoi(round)
	} else {
		a.round = 2
	}
	if tag != "" && date != "" {
		dateTime, err := a.excelTimeToTime(date)
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		tagValue_f, err := a.store.GetTagDate(tag, dateTime)
		if err != nil {
			w.Write([]byte("#Error: " + err.Error()))
			return
		}
		if format == "raw" {
			w.Write([]byte(fmt.Sprintf("%f", tagValue_f)))
		} else {
			w.Write([]byte(a.store.RoundAndFormat(tagValue_f)))
		}
		return
	} else if tag != "" && from != "" && to != "" {
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
		var tagValue string
		if count != "" {
			countT, err := strconv.Atoi(count)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			v, err := a.store.GetTagCount(tag, fromT, toT, countT)
			if err != nil {
				w.Write([]byte("#Error: " + err.Error()))
				return
			} else {
				tagValue, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					w.Write([]byte("#Error: " + err.Error()))
					return
				}
				w.Write([]byte(tagValue))
				return
			}
		} else {
			if group == "" {
				v, err := a.store.GetTagFromTo(tag, fromT, toT)
				if err != nil {
					w.Write([]byte("#Error: " + err.Error()))
					return
				} else {
					tagValue, err := json.MarshalIndent(v, "", "  ")
					if err != nil {
						w.Write([]byte("#Error: " + err.Error()))
						return
					}
					w.Write([]byte(tagValue))
					return
				}
			} else {
				tagValue = a.store.GetTagFromToGroup(tag, fromT, toT, group)
			}
		}
		w.Write([]byte(tagValue))
	}
}

// send file "tests_book.xls" to client
// func (a *App) handleGetExample(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
// 	w.Header().Set("Content-Type", "application/vnd.ms-excel")
// 	w.Header().Set("Content-Disposition", "attachment;filename=\"tests_book.xlsx\"")
// 	file, _ := os.Open("../config/tests_book.xlsx")
// 	defer func() {
// 		err := file.Close()
// 		if err != nil {
// 			logger.Error(err.Error())
// 		}
// 	}()
// 	io.Copy(w, file)
// }

// returns JSON with tags by mask
func (a *App) handleGetTagList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	like := r.URL.Query().Get("like")
	tags, err := a.store.GetTagList(like)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	j, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		// logger.Debug(err.error())
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
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

// func (a *App) tryParseDateUTC(date string) (time.Time, error) {
// 	// if date is empty, return error
// 	if date == "" {
// 		return time.Time{}, errors.ErrInvalidDate
// 	}
// 	// if date is not empty, try to parse it to time.Time
// 	// if date is not valid, return error
// 	cfg := a.config.GetStringSlice("app.date_formats")
// 	for fm := range cfg {
// 		t, err := time.ParseInLocation(cfg[fm], date, time.UTC)
// 		if err == nil {
// 			return t, nil
// 		}
// 	}
// 	return time.Time{}, errors.ErrInvalidDate
// }

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

// output program log
func (a *App) handleGetLog(w http.ResponseWriter, r *http.Request) {
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

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "text/html")
	// get template file
	t, err := template.ParseFiles(a.workDir + "/web/templates/index.html")
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	type Page struct {
		Name    string
		Version string
	}
	// execute template
	t.Execute(w, Page{a.name, a.version})
	// f, _ := os.ReadFile(a.workDir + "/web/index.html")
	// w.Write(f)
}

func (a *App) handleScripts(w http.ResponseWriter, r *http.Request) {
	basePath := filepath.Join(a.workDir, "web", "scripts")
	filePath := filepath.Join(basePath, r.URL.Path[len("/images/"):])
	if !strings.HasPrefix(filePath, basePath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	logger.Debug(filePath)
	http.ServeFile(w, r, filePath)
}

func (a *App) handleImages(w http.ResponseWriter, r *http.Request) {
	basePath := filepath.Join(a.workDir, "web", "images")
	filePath := filepath.Join(basePath, r.URL.Path[len("/images/"):])
	if !strings.HasPrefix(filePath, basePath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	logger.Debug(filePath)
	http.ServeFile(w, r, filePath)
}
