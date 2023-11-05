package robin

import (
	"fmt"
	"net/http"
	"robin2/internal/format"
	"robin2/internal/logger"
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
	logger.Trace("list templates")
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
	logger.Trace("adding template")
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
	logger.Trace("getting template")
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
	logger.Trace("editing template")
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
	logger.Trace("deleting template")
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
// @Param db query string false "Имя базы данных"
// @Param format query string false "Формат вывода (str - по умолчанию, json, raw)"
// @Param args query array false "Список аргументов"
// @x-try-it-out-enabled false
func (a *App) handleTemplateExec(w http.ResponseWriter, r *http.Request) {
	logger.Trace("executing template")
	writer := []byte("#Error: unknown error")
	defer func() {
		w.Write(writer)
	}()
	name := r.URL.Query().Get("name")
	if name == "" {
		writer = []byte("#Error: name is empty")
		return
	}

	formatStr := r.URL.Query().Get("format")
	params := make(map[string]string)
	args := r.URL.Query().Get("args")
	for _, arg := range strings.Split(args, ",") {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue
		}
		params[kv[0]] = kv[1]
	}

	db := r.URL.Query().Get("db")
	params["db"] = db

	b, err := a.store.TemplateExec(name, params)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}

	writer = format.New(formatStr).Process(b)
}
