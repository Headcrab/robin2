// todo: sessions in web api to divide users data caches
package robin

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"robin2/internal/data"
	"robin2/internal/logger"
	"robin2/internal/utils"
	"strconv"
	"strings"
	"time"
	// swagger "github.com/swaggo/http-swagger/v2"
)

func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, a.workDir+"/web/images/icon.png")
}

func (a *App) handleDirectory(d string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

func getOnePage(name string, descr string, data []string, pageNum, linesPerPage int) map[string]interface{} {
	pagesTotal := len(data)/(linesPerPage+1) + 1
	// pagesTotal = thenIf(pagesTotal == 0, 1, pagesTotal)
	if pageNum > pagesTotal {
		pageNum = pagesTotal
	}

	pageSwicher := generatePageSwitcherHTML(name, pageNum, pagesTotal)

	dataL := getDataSubset(data, pageNum, linesPerPage)

	return map[string]interface{}{
		name:    dataL,
		"descr": descr,
		"page":  pageSwicher,
	}
}

// generatePageSwitcherHTML generates the HTML for the page switcher component.
func generatePageSwitcherHTML(name string, pageNum, pagesTotal int) template.HTML {
	if pagesTotal < 2 {
		return template.HTML("")
	}

	pageSwitcher := "<div class='flex items-center justify-center gap-1'>"

	// Previous button
	if pageNum > 1 {
		pageSwitcher += getFormattedPageNumber(name, pageNum-1, false, "‹")
	}

	// First page
	pageSwitcher += getFormattedPageNumber(name, 1, pageNum == 1, "")

	pageLimit := 11
	if pagesTotal <= pageLimit {
		// Show all pages if total is small
		for i := 2; i < pagesTotal; i++ {
			pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
		}
	} else {
		// Show ellipsis and middle pages for large pagination
		if pageNum <= 4 {
			// Show pages 2-5 and ellipsis
			for i := 2; i <= 5; i++ {
				if i < pagesTotal {
					pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
				}
			}
			if pagesTotal > 6 {
				pageSwitcher += "<span class='page-ellipsis'>...</span>"
			}
		} else if pageNum >= pagesTotal-3 {
			// Show ellipsis and last 4 pages
			if pagesTotal > 6 {
				pageSwitcher += "<span class='page-ellipsis'>...</span>"
			}
			for i := pagesTotal - 4; i < pagesTotal; i++ {
				if i > 1 {
					pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
				}
			}
		} else {
			// Show ellipsis, middle pages, ellipsis
			pageSwitcher += "<span class='page-ellipsis'>...</span>"
			for i := pageNum - 1; i <= pageNum+1; i++ {
				pageSwitcher += getFormattedPageNumber(name, i, pageNum == i, "")
			}
			pageSwitcher += "<span class='page-ellipsis'>...</span>"
		}
	}

	// Last page (if not already shown)
	if pagesTotal > 1 {
		pageSwitcher += getFormattedPageNumber(name, pagesTotal, pageNum == pagesTotal, "")
	}

	// Next button
	if pageNum < pagesTotal {
		pageSwitcher += getFormattedPageNumber(name, pageNum+1, false, "›")
	}

	pageSwitcher += "</div>"
	return template.HTML(pageSwitcher)
}

func getFormattedPageNumber(name string, pageNum int, isCurr bool, pagerName string) string {
	currentClass := ""
	if isCurr {
		currentClass = " page-number-current"
	}

	displayText := pagerName
	if pagerName == "" {
		displayText = fmt.Sprintf("%d", pageNum)
	}

	return fmt.Sprintf(`<button class='page-number%s' onclick='loadPage("/%s?page=%d")' type='button'>%s</button>`,
		currentClass, name, pageNum, displayText)
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
		dataFull := map[string]interface{}{
			"descr":   data["descr"],
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
	// procTimeBegin := time.Now()
	page := "logs"
	logPerPage := 23
	pageNumStr := r.URL.Query().Get("page")
	refresh := r.URL.Query().Get("refresh")

	if pageNumStr == "" {
		pageNumStr = "1"
		logData = nil
	}

	// если указан параметр refresh, очищаем кеш
	if refresh == "1" {
		logData = nil
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if logData == nil {
		logs, err := logger.GetLogHistory()
		if err != nil {
			fmt.Println("Ошибка при чтении ответа:", err)
			return
		}
		logData = []string{} // инициализируем пустой слайс
		for _, log := range logs {
			logData = append(logData, fmt.Sprintf("%s %s %s", log.Date.Format("2006-01-02 15:04:05"), log.Level, log.Msg))
		}
	}
	a.handlePageAny(page, getOnePage(page, "Лог", logData, pageNum, logPerPage))(w, r)

}

var tagsValues data.Tags

func (a *App) handlePageData(w http.ResponseWriter, r *http.Request) {
	// procTimeBegin := time.Now()
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
			tagsValues, err = a.store.GetTagCountGroup(q.Get("tag"), from, to, count, "avg")
			if err != nil {
				fmt.Println("Ошибка при чтении ответа:", err)
				return
			}
		}
	}

	data := []string{}
	for _, tag := range tagsValues {
		data = append(data, fmt.Sprintf("%s|%f", tag.Date, tag.Value))
	}
	a.handlePageAny(page, getOnePage(page, "Получение данных", data, pageNum, linesPerPage))(w, r)

}

var tagsList []string

func (a *App) handlePageTags(w http.ResponseWriter, r *http.Request) {
	// procTimeBegin := time.Now()
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
				_, err := w.Write([]byte("#Error: " + err.Error()))
				if err != nil {
					logger.Error(fmt.Sprintf("Error writing response: %v", err))
				}
				return
			}
			for _, tag := range tags.Rows {
				tagsList = append(tagsList, tag[0])
			}
		}
	}

	data := getOnePage(page, "Тэги", tagsList, pageNum, linesPerPage)
	data["like"] = utils.ThenIf(like == "", "", like)
	a.handlePageAny(page, data)(w, r)
}

func (a *App) handlePageSwagger(w http.ResponseWriter, r *http.Request) {
	// get data from /swagger
	page := "swagger"
	data := map[string]interface{}{
		"descr": "Документация API",
		"name":  "swagger",
		// "content": "string",
	}

	a.handlePageAny(page, data)(w, r)

}
