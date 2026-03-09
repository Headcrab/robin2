package robin

import (
	"fmt"
	"net/http"
	"robin2/internal/format"
	"robin2/internal/logger"
	"strings"
)

// @Summary Получить список шаблонов
// @Description Возвращает список шаблонов. Требует admin token
// @Tags Template
// @Produce plain
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Router /templ/list/ [get]
// @Param like query string false "Маска поиска шаблона"
func (a *App) handleTemplateList(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("list templates")
	like := r.URL.Query().Get("like")
	if !isSafeTemplateLike(like) {
		http.Error(w, "invalid template mask", http.StatusBadRequest)
		return
	}

	b, err := a.store.TemplateList(like)
	if err != nil {
		writeStringResponse(w, "#Error: "+err.Error())
		return
	}

	res := fmt.Sprintf("Templates like %s (%v)\n\n ", like, len(b))
	for k, v := range b {
		res += k + "\n " + v + "\n\n"
	}
	writeStringResponse(w, res)
}

// @Summary Добавить шаблон
// @Description Добавляет шаблон. Требует admin token
// @Tags Template
// @Produce plain
// @Accept x-www-form-urlencoded
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 405 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Param name formData string true "Имя шаблона"
// @Param body formData string true "Тело шаблона"
// @Router /templ/add/ [post]
func (a *App) handleTemplateAdd(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("adding template")
	name := r.FormValue("name")
	if name == "" {
		writeStringResponse(w, "#Error: name is empty")
		return
	}
	if !isSafeTemplateName(name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	body := r.FormValue("body")
	if body == "" {
		writeStringResponse(w, "#Error: body is empty")
		return
	}

	err := a.store.TemplateAdd(name, body)
	if err != nil {
		writeStringResponse(w, "#Error: "+err.Error())
		return
	}

	writeStringResponse(w, fmt.Sprintf("Template %s added", name))
}

// @Summary Получить тело шаблона
// @Description Возвращает тело шаблона. Требует admin token
// @Tags Template
// @Produce plain
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Router /templ/get/ [get]
// @Param name query string true "Имя шаблона"
func (a *App) handleTemplateGet(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("getting template")
	name := r.URL.Query().Get("name")
	if name == "" {
		writeStringResponse(w, "#Error: name is empty")
		return
	}
	if !isSafeTemplateName(name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	b, err := a.store.TemplateGet(name)
	if err != nil {
		writeStringResponse(w, "#Error: "+err.Error())
		return
	}
	writeStringResponse(w, b)
}

// @Summary Изменить тело шаблона
// @Description Изменяет тело шаблона. Требует admin token
// @Tags Template
// @Produce plain
// @Accept x-www-form-urlencoded
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 405 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Param name formData string true "Имя шаблона"
// @Param body formData string true "Тело шаблона"
// @Router /templ/edit/ [post]
func (a *App) handleTemplateEdit(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("editing template")
	name := r.FormValue("name")
	if name == "" {
		writeStringResponse(w, "#Error: name is empty")
		return
	}
	if !isSafeTemplateName(name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	body := r.FormValue("body")
	if body == "" {
		writeStringResponse(w, "#Error: body is empty")
		return
	}

	err := a.store.TemplateSet(name, body)
	if err != nil {
		writeStringResponse(w, "#Error: "+err.Error())
		return
	}

	writeStringResponse(w, fmt.Sprintf("Template %s edited", name))
}

// @Summary Удалить шаблон
// @Description Удаляет шаблон. Требует admin token
// @Tags Template
// @Produce plain
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 405 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Param name query string true "Имя шаблона"
// @Router /templ/delete/ [delete]
func (a *App) handleTemplateDelete(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("deleting template")
	name := r.FormValue("name")
	if name == "" {
		writeStringResponse(w, "#Error: name is empty")
		return
	}
	if !isSafeTemplateName(name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	err := a.store.TemplateDel(name)
	if err != nil {
		writeStringResponse(w, "#Error: "+err.Error())
		return
	}

	writeStringResponse(w, fmt.Sprintf("Template %s deleted", name))
}

// @Summary Выполнить шаблон
// @Description Выполняет шаблон. Требует admin token
// @Tags Template
// @Produce plain
// @Accept x-www-form-urlencoded
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Failure 405 {string} string
// @Param X-Admin-Token header string false "Admin token"
// @Param Authorization header string false "Bearer admin token"
// @Param name formData string true "Имя шаблона"
// @Param db formData string false "Имя базы данных"
// @Param format formData string false "Формат вывода (text, str, raw, json, xml, html, grafana)"
// @Param args formData string false "Список аргументов k1=v1,k2=v2"
// @x-try-it-out-enabled false
// @Router /templ/exec/ [post]
func (a *App) handleTemplateExec(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}

	logger.Trace("executing template")
	writer := []byte("#Error: unknown error")
	defer func() {
		writeResponse(w, writer)
	}()
	name := r.FormValue("name")
	if name == "" {
		writer = []byte("#Error: name is empty")
		return
	}
	if !isSafeTemplateName(name) {
		http.Error(w, "invalid template name", http.StatusBadRequest)
		return
	}

	formatStr := r.FormValue("format")
	params := make(map[string]string)
	args := r.FormValue("args")
	for _, arg := range strings.Split(args, ",") {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue
		}
		if !isSafeTemplateArgKey(kv[0]) {
			http.Error(w, "invalid template argument key", http.StatusBadRequest)
			return
		}
		params[kv[0]] = kv[1]
	}

	db := r.FormValue("db")
	if db != "" && !isSafeTemplateName(db) {
		http.Error(w, "invalid database name", http.StatusBadRequest)
		return
	}
	params["db"] = db

	b, err := a.store.TemplateExec(name, params)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	fmtr, err := format.New(formatStr)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}
	writer = fmtr.Process(b)
}
