package robin

import (
	"encoding/json"
	"fmt"
	"net"

	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"robin2/internal/cache"
	"robin2/internal/errors"
	"robin2/internal/store"
	"robin2/pkg/config"
	"robin2/pkg/logger"
)

func NewApp() *App {
	app := &App{
		name:    "Robin",
		version: "2.0.0",
	}
	return app
}

type App struct {
	name      string
	version   string
	config    config.Config
	cache     cache.BaseCache
	startTime time.Time
	round     int
	opCount   int64
	store     store.BaseStore
}

func (a *App) Run() {
	a.startTime = time.Now()
	logger.Log(logger.Info, a.name+" "+a.version+" is running")
	a.init()
	err := a.store.Connect(a.cache)
	if err != nil {
		logger.Log(logger.Fatal, err.Error())
	}
	for _, ip := range getLocalhostIpAdresses() {
		logger.Log(logger.Info, "listening on http://"+ip+":"+a.config.GetString("app.port"))
	}
	err = http.ListenAndServe(":"+a.config.GetString("app.port"), nil)
	if err != nil {
		logger.Log(logger.Fatal, err.Error())
	}
}

func getLocalhostIpAdresses() []string {
	var loclocalhostIpAdresses []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Log(logger.Error, err.Error())
	}
	loclocalhostIpAdresses = append(loclocalhostIpAdresses, "127.0.0.1")
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				loclocalhostIpAdresses = append(loclocalhostIpAdresses, ipnet.IP.String())
			}
		}
	}
	return loclocalhostIpAdresses
}

func (a *App) init() {
	logger.Log(logger.Debug, "initializing app")
	a.config = *config.GetConfig()
	a.cache = cache.NewCacheFactory().NewCache(a.config.GetString("app.cache.type"))
	a.store = *store.NewStoreFactory().NewStore(a.config.GetString("app.db.type"))
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		// "/robin/":               a.handleHome,
		// "/info/":   a.handleInfo,
		// "/uptime/":        a.handleUptime,
		// "/reload_config/": a.handleReloadConfig,
		// "/get/example/":   a.handleGetExample,
		"/get/tag/":      a.handleGetTag,
		"/get/tag/list/": a.handleGetTagList,
		// "/favicon.ico":    a.handleFavicon,
	}
	for path, handler := range handlers {
		http.HandleFunc(path, handler)
	}
}

func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	logger.Log(logger.Debug, "favicon")
	http.ServeFile(w, r, "../website/favicon.ico")
}

func (a *App) handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	logger.Log(logger.Info, "reloading config file")
	err := a.config.Reload()
	if err != nil {
		logger.Log(logger.Fatal, "Failed to read config file")
	}
	w.Write([]byte("ok"))

}

func (a *App) handleUptime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	uptime := time.Since(a.startTime).Round(time.Second)
	_, err := w.Write([]byte(uptime.String()))
	if err != nil {
		logger.Log(logger.Error, err.Error())
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

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	f, _ := os.ReadFile("index.html")
	w.Write(f)
}

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
func (a *App) handleGetExample(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Header().Set("Content-Disposition", "attachment;filename=\"tests_book.xlsx\"")
	file, _ := os.Open("../config/tests_book.xlsx")
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Log(logger.Error, err.Error())
		}
	}()
	io.Copy(w, file)
}

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
		// logger.Log(logger.Debug, err.error())
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	w.Write(j)
}

// return time.Time and error
func (a *App) excelTimeToTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, errors.ErrEmptyDate
	}
	var result = time.Time{}
	if !strings.Contains(timeStr, ":") {
		timeStr = strings.Replace(timeStr, ",", ".", 1)
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return time.Time{}, errors.ErrNotAFloat
		}
		timeInSecs := int64(math.Round(timeFloat * 24 * 60 * 60))
		result = time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local).Add(time.Duration(timeInSecs) * time.Second)
		result.In(time.Local)
	} else {
		res, err := a.tryParseDate(timeStr)
		if err != nil {
			res = time.Time{}
			// logger.Log(logger.Warn, err.logger.Error())
			return time.Time{}, err
		}
		result = time.Date(res.Year(), res.Month(), res.Day(), res.Hour(), res.Minute(), res.Second(), 0, time.Local)
	}
	if result.IsZero() {
		return time.Time{}, errors.ErrEmptyDate
	}
	return result, nil
}

func (a *App) tryParseDate(date string) (time.Time, error) {
	// if date is empty, return error
	if date == "" {
		return time.Time{}, errors.ErrEmptyDate
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
