// todo: sessions in web api to divide users data caches
package robin

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"robin2/internal/logger"
	"robin2/internal/utils"
	"sort"
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
		utils.ThenIf(isCurr, "page-number-current", ""), name, pageNum,
		utils.ThenIf(pagerName == "", fmt.Sprintf("%d", pageNum), pagerName))
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
	a.handlePageAny(page, getOnePage(page, logData, pageNum, logPerPage))(w, r)

}

var tagsValues map[string]map[time.Time]float32

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
		times := make([]time.Time, 0, len(tag))
		for k := range tag {
			times = append(times, k)
		}
		sort.Slice(times, func(i, j int) bool {
			return times[i].Before(times[j])
		})
		for _, v := range times {
			data = append(data, fmt.Sprintf("%s|%f", v, tag[v]))
		}
	}
	a.handlePageAny(page, getOnePage(page, data, pageNum, linesPerPage))(w, r)

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
				w.Write([]byte("#Error: " + err.Error()))
				return
			}
			for _, tag := range tags.Rows {
				tagsList = append(tagsList, tag[0])
			}
		}
	}

	data := getOnePage(page, tagsList, pageNum, linesPerPage)
	data["like"] = utils.ThenIf(like == "", "", like)
	a.handlePageAny(page, data)(w, r)
}

func (a *App) handlePageSwagger(w http.ResponseWriter, r *http.Request) {
	// get data from /swagger
	page := "swagger"
	data := map[string]interface{}{
		"name":    "swagger",
		"content": "string",
	}

	a.handlePageAny(page, data)(w, r)

}
