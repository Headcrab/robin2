// todo: sessions in web api to divide users data caches
package robin

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"robin2/internal/logger"
	"robin2/internal/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
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
	pagesTotal := 1
	if linesPerPage > 0 && len(data) > 0 {
		pagesTotal = (len(data) + linesPerPage - 1) / linesPerPage
	}
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
		return ""
	}

	pages := getVisiblePageNumbers(pageNum, pagesTotal)
	parts := make([]string, 0, len(pages)+2)
	prevPage := 0
	for _, p := range pages {
		if prevPage != 0 && p-prevPage > 1 {
			parts = append(parts, `<span class="pagination-ellipsis" aria-hidden="true">...</span>`)
		}
		parts = append(parts, getFormattedPageNumber(name, p, pageNum == p, ""))
		prevPage = p
	}

	return template.HTML(strings.Join(parts, ""))
}

func getFormattedPageNumber(name string, pageNum int, isCurr bool, pagerName string) string {
	label := utils.ThenIf(pagerName == "", fmt.Sprintf("%d", pageNum), pagerName)
	url := fmt.Sprintf("/%s?page=%d", name, pageNum)
	if isCurr {
		return fmt.Sprintf(`<span class="active" aria-current="page">%s</span>`, label)
	}
	return fmt.Sprintf(`<a href="%s" onclick='loadPage("%s"); return false;'>%s</a>`, url, url, label)
}

func getVisiblePageNumbers(pageNum, pagesTotal int) []int {
	pageSet := map[int]struct{}{
		1:          {},
		pagesTotal: {},
		pageNum:    {},
	}

	for i := pageNum - 1; i <= pageNum+1; i++ {
		if i > 1 && i < pagesTotal {
			pageSet[i] = struct{}{}
		}
	}

	if pageNum <= 4 {
		for i := 2; i <= min(5, pagesTotal-1); i++ {
			pageSet[i] = struct{}{}
		}
	}

	if pageNum >= pagesTotal-3 {
		for i := max(2, pagesTotal-4); i < pagesTotal; i++ {
			pageSet[i] = struct{}{}
		}
	}

	pages := make([]int, 0, len(pageSet))
	for p := range pageSet {
		pages = append(pages, p)
	}
	sort.Ints(pages)
	return pages
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

func (a *App) handlePageLog(w http.ResponseWriter, r *http.Request) {
	a.pageCache.mu.Lock()
	defer a.pageCache.mu.Unlock()

	// procTimeBegin := time.Now()
	page := "logs"
	logPerPage := 23
	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		pageNumStr = "1"
		a.pageCache.logData = nil
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if a.pageCache.logData == nil {
		logs, err := logger.GetLogHistory()
		if err != nil {
			fmt.Println("Ошибка при чтении ответа:", err)
			return
		}
		for _, log := range logs {
			a.pageCache.logData = append(a.pageCache.logData, fmt.Sprintf("%s %s %s", log.Date.Format("2006-01-02 15:04:05"), log.Level, log.Msg))
		}
	}
	a.handlePageAny(page, getOnePage(page, "Лог", a.pageCache.logData, pageNum, logPerPage))(w, r)

}

func (a *App) handlePageData(w http.ResponseWriter, r *http.Request) {
	a.pageCache.mu.Lock()
	defer a.pageCache.mu.Unlock()

	// procTimeBegin := time.Now()
	page := "data"

	linesPerPage := 23

	q := r.URL.Query()
	pageNumStr := q.Get("page")
	if pageNumStr == "" {
		a.pageCache.tagsValues = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	if a.pageCache.tagsValues == nil {
		if q.Get("tag") != "" && q.Get("from") != "" && q.Get("to") != "" {
			from, _ := time.Parse("2006-01-02T15:04", q.Get("from"))
			to, _ := time.Parse("2006-01-02T15:04", q.Get("to"))
			countStr, _ := strconv.Atoi(q.Get("count"))
			count := int(countStr)
			// tags, err = a.store.GetTagFromTo(q.Get("tag"), from, to)
			a.pageCache.tagsValues, err = a.store.GetTagCountGroup(q.Get("tag"), from, to, count, "avg")
			if err != nil {
				fmt.Println("Ошибка при чтении ответа:", err)
				return
			}
		}
	}

	data := []string{}
	for _, tag := range a.pageCache.tagsValues {
		data = append(data, fmt.Sprintf("%s|%f", tag.Date, tag.Value))
	}
	a.handlePageAny(page, getOnePage(page, "Получение данных", data, pageNum, linesPerPage))(w, r)

}

func (a *App) handlePageTags(w http.ResponseWriter, r *http.Request) {
	a.pageCache.mu.Lock()
	defer a.pageCache.mu.Unlock()

	// procTimeBegin := time.Now()
	page := "tags"
	linesPerPage := 23

	pageNumStr := r.URL.Query().Get("page")
	if pageNumStr == "" {
		a.pageCache.tagsList = nil
		pageNumStr = "1"
	}

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	like := r.URL.Query().Get("like")
	if like != "" {
		if a.pageCache.tagsList == nil {
			tags, err := a.store.GetTagList(like)
			if err != nil {
				_, err := w.Write([]byte("#Error: " + err.Error()))
				if err != nil {
					logger.Error(fmt.Sprintf("Error writing response: %v", err))
				}
				return
			}
			for _, tag := range tags.Rows {
				a.pageCache.tagsList = append(a.pageCache.tagsList, tag[0])
			}
		}
	}

	data := getOnePage(page, "Тэги", a.pageCache.tagsList, pageNum, linesPerPage)
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

// handlePageDocs handles /docs/ requests and displays list of markdown files
func (a *App) handlePageDocs(w http.ResponseWriter, r *http.Request) {
	page := "docs"

	// Get list of markdown files from docs folder
	docFiles, err := a.getMarkdownFiles()
	if err != nil {
		logger.Error("Error reading docs folder: " + err.Error())
		http.Error(w, "Error reading documentation files", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"descr": "Документация",
		"docs":  docFiles,
	}

	a.handlePageAny(page, data)(w, r)
}

// getMarkdownFiles returns list of markdown files from docs folder
func (a *App) getMarkdownFiles() ([]map[string]interface{}, error) {
	docsPath := filepath.Join(a.workDir, "docs")
	files, err := os.ReadDir(docsPath)
	if err != nil {
		return nil, err
	}

	var docFiles []map[string]interface{}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			docFiles = append(docFiles, map[string]interface{}{
				"name":     file.Name(),
				"title":    strings.TrimSuffix(file.Name(), ".md"),
				"size":     info.Size(),
				"modified": info.ModTime(),
			})
		}
	}

	return docFiles, nil
}

// handleDocView handles /docs/view/ requests and displays specific markdown file
func (a *App) handleDocView(w http.ResponseWriter, r *http.Request) {
	page := "doc-view"
	fileName := r.URL.Query().Get("file")

	if fileName == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	// Security check - only allow .md files and prevent path traversal
	if !strings.HasSuffix(strings.ToLower(fileName), ".md") {
		http.Error(w, "Only markdown files are allowed", http.StatusBadRequest)
		return
	}

	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	// Read markdown file
	filePath := filepath.Join(a.workDir, "docs", fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Error reading doc file: " + err.Error())
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Render markdown to HTML
	htmlContent, err := a.renderMarkdown(content)
	if err != nil {
		logger.Error("Error rendering markdown: " + err.Error())
		http.Error(w, "Error processing document", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"descr":    "Документ: " + strings.TrimSuffix(fileName, ".md"),
		"filename": fileName,
		"title":    strings.TrimSuffix(fileName, ".md"),
		"content":  htmlContent,
	}

	a.handlePageAny(page, data)(w, r)
}

// renderMarkdown converts markdown content to HTML using goldmark
func (a *App) renderMarkdown(content []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(content, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
