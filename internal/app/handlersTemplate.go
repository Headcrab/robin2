package robin

import (
	"fmt"
	"net/http"
	"robin2/pkg/logger"
	"strings"
)

// @Summary Получить список шаблонов
// @Description Возвращает список шаблонов
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/list [get]
// @Param like query string false "Маска поиска шаблона"
func (a *App) handleTemplateList(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	like := r.URL.Query().Get("like")

	b, err := a.store.TemplateList(like)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	res := fmt.Sprintf("Templates like %s (%v)\n\n ", like, len(b))
	for k, v := range b {
		res += k + "\n " + v + "\n\n"
	}
	w.Write([]byte(res))
}

// @Summary Добавить шаблон
// @Description Добавляет шаблон
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/add [get]
// @Param name query string true "Имя шаблона"
// @Param body query string true "Тело шаблона"
func (a *App) handleTemplateAdd(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name == "" {
		w.Write([]byte("#Error: name is empty"))
	}

	body := r.URL.Query().Get("body")
	if body == "" {
		w.Write([]byte("#Error: body is empty"))
		return
	}

	err := a.store.TemplateAdd(name, body)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
	}

	w.Write([]byte(fmt.Sprintf("Template %s added", name)))
}

// @Summary Получить тело шаблона
// @Description Возвращает тело шаблона
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/get [get]
// @Param name query string true "Имя шаблона"
func (a *App) handleTemplateGet(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name == "" {
		w.Write([]byte("#Error: name is empty"))
		return
	}

	b, err := a.store.TemplateGet(name)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	w.Write([]byte(b))
}

// @Summary Изменить тело шаблона
// @Description Изменяет тело шаблона
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/edit [get]
// @Param name query string true "Имя шаблона"
// @Param body query string true "Тело шаблона"
func (a *App) handleTemplateEdit(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name == "" {
		w.Write([]byte("#Error: name is empty"))
		return
	}

	body := r.URL.Query().Get("body")
	if body == "" {
		w.Write([]byte("#Error: body is empty"))
		return
	}

	err := a.store.TemplateSet(name, body)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("Template %s edited", name)))
}

// @Summary Удалить шаблон
// @Description Удаляет шаблон
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/del [get]
// @Param name query string true "Имя шаблона"
func (a *App) handleTemplateDelete(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")

	name := r.URL.Query().Get("name")
	if name == "" {
		w.Write([]byte("#Error: name is empty"))
		return
	}

	err := a.store.TemplateDel(name)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("Template %s deleted", name)))
}

// @Summary Выполнить шаблон
// @Description Выполняет шаблон
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/exec [get]
// @Param name query string true "Имя шаблона"
// @Param args query array false "Список аргументов"
// @x-try-it-out-enabled false
func (a *App) handleTemplateExec(w http.ResponseWriter, r *http.Request) {
	logger.Trace("rendered template page")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	if name == "" {
		w.Write([]byte("#Error: name is empty"))
		return
	}

	params := make(map[string]string)
	args := r.URL.Query().Get("args")
	for _, arg := range strings.Split(args, ",") {
		kv := strings.Split(arg, "=")
		params[kv[0]] = kv[1]
	}

	b, err := a.store.TemplateExec(name, params)
	if err != nil {
		w.Write([]byte("#Error: " + err.Error()))
		return
	}
	w.Write([]byte(b))
}