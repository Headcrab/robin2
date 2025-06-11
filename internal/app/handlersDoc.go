package robin

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"robin2/internal/logger"
	"strings"
)

type Document struct {
	Name        string
	DisplayName string
	Description string
	Size        string
	Content     template.HTML
}

// initDocuments инициализирует список документов при запуске приложения
func (a *App) initDocuments() []Document {
	var docs []Document
	docsDir := filepath.Join(a.workDir, "docs")

	// читаем все .md файлы из папки docs
	files, err := ioutil.ReadDir(docsDir)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read docs directory: %v", err))
		return docs
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			doc := Document{
				Name:        file.Name(),
				DisplayName: getDisplayName(file.Name()),
				Description: getDocDescription(filepath.Join(docsDir, file.Name())),
				Size:        formatFileSize(file.Size()),
			}
			docs = append(docs, doc)
		}
	}

	logger.Info(fmt.Sprintf("Loaded %d documents", len(docs)))
	return docs
}

// handlePageDocs обрабатывает запрос на страницу документации
func (a *App) handlePageDocs(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered docs page")

	// получаем список документов при каждом запросе для актуальности
	docs := a.initDocuments()

	data := map[string]interface{}{
		"descr": "Документация",
		"docs":  docs,
	}

	a.handlePageAny("docs", data)(w, r)
}

// handlePageDocView обрабатывает запрос на просмотр конкретного документа
func (a *App) handlePageDocView(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	// проверяем безопасность пути
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	docsDir := filepath.Join(a.workDir, "docs")
	filePath := filepath.Join(docsDir, fileName)

	// проверяем что файл существует и является .md
	if !strings.HasSuffix(strings.ToLower(fileName), ".md") {
		http.Error(w, "Only markdown files are allowed", http.StatusBadRequest)
		return
	}

	// читаем содержимое файла
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read file %s: %v", fileName, err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// конвертируем markdown в HTML (простая реализация)
	htmlContent := convertMarkdownToHTML(string(content))

	// для отладки логируем первые 200 символов
	if len(content) > 200 {
		logger.Trace(fmt.Sprintf("Original markdown (first 200 chars): %s", string(content)[:200]))
	} else {
		logger.Trace(fmt.Sprintf("Original markdown: %s", string(content)))
	}
	if len(htmlContent) > 200 {
		logger.Trace(fmt.Sprintf("Converted HTML (first 200 chars): %s", htmlContent[:200]))
	} else {
		logger.Trace(fmt.Sprintf("Converted HTML: %s", htmlContent))
	}

	data := map[string]interface{}{
		"descr":       "Просмотр документа",
		"title":       getDisplayName(fileName),
		"description": "Документ: " + fileName,
		"content":     template.HTML(htmlContent),
	}

	a.handlePageAny("doc-view", data)(w, r)
}

// getDisplayName получает человекочитаемое имя документа
func getDisplayName(fileName string) string {
	name := strings.TrimSuffix(fileName, ".md")

	// заменяем подчеркивания и дефисы на пробелы
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// делаем первую букву заглавной
	if len(name) > 0 {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}

	return name
}

// getDocDescription получает описание документа из первой строки
func getDocDescription(filePath string) string {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "Документ"
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// убираем markdown разметку
			line = strings.ReplaceAll(line, "*", "")
			line = strings.ReplaceAll(line, "_", "")
			line = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(line, "$1")

			if len(line) > 100 {
				line = line[:97] + "..."
			}
			return line
		}
		if strings.HasPrefix(line, "# ") {
			return "Документация проекта"
		}
	}

	return "Документ"
}

// formatFileSize форматирует размер файла в человекочитаемом виде
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// convertMarkdownToHTML простая конверсия markdown в HTML
func convertMarkdownToHTML(markdown string) string {
	html := markdown

	// экранируем HTML символы
	html = strings.ReplaceAll(html, "&", "&amp;")
	html = strings.ReplaceAll(html, "<", "&lt;")
	html = strings.ReplaceAll(html, ">", "&gt;")

	// заголовки (порядок важен - сначала более специфичные)
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")

	// жирный текст
	html = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(html, "<strong>$1</strong>")
	html = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(html, "<strong>$1</strong>")

	// курсив (после жирного текста)
	html = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(html, "<em>$1</em>")
	html = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(html, "<em>$1</em>")

	// код блоки (сначала многострочные)
	html = regexp.MustCompile("```([\\s\\S]*?)```").ReplaceAllString(html, "<pre><code>$1</code></pre>")
	// затем однострочные
	html = regexp.MustCompile("`([^`]+)`").ReplaceAllString(html, "<code>$1</code>")

	// ссылки
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, `<a href="$2" target="_blank">$1</a>`)

	// обработка списков построчно
	lines := strings.Split(html, "\n")
	var result []string
	inList := false
	listType := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// пропускаем пустые строки
		if trimmed == "" {
			continue
		}

		// маркированные списки
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList || listType != "ul" {
				if inList {
					result = append(result, fmt.Sprintf("</%s>", listType))
				}
				result = append(result, "<ul>")
				inList = true
				listType = "ul"
			}
			content := strings.TrimSpace(trimmed[2:])
			result = append(result, fmt.Sprintf("<li>%s</li>", content))
		} else if regexp.MustCompile(`^\d+\. `).MatchString(trimmed) {
			// нумерованные списки
			if !inList || listType != "ol" {
				if inList {
					result = append(result, fmt.Sprintf("</%s>", listType))
				}
				result = append(result, "<ol>")
				inList = true
				listType = "ol"
			}
			content := regexp.MustCompile(`^\d+\. `).ReplaceAllString(trimmed, "")
			result = append(result, fmt.Sprintf("<li>%s</li>", content))
		} else {
			// закрываем список если нужно
			if inList {
				result = append(result, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}

			// обычные параграфы или заголовки
			if !strings.HasPrefix(trimmed, "<h") {
				result = append(result, fmt.Sprintf("<p>%s</p>", trimmed))
			} else {
				result = append(result, trimmed)
			}
		}
	}

	// закрываем список в конце если нужно
	if inList {
		result = append(result, fmt.Sprintf("</%s>", listType))
	}

	return strings.Join(result, "")
}
