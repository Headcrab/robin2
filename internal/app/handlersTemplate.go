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
		if _, err := w.Write([]byte("#Error: " + err.Error())); err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	res := fmt.Sprintf("Templates like %s (%v)\n\n ", like, len(b))
	for k, v := range b {
		res += k + "\n " + v + "\n\n"
	}
	if _, err := w.Write([]byte(res)); err != nil {
		logger.Error(fmt.Sprintf("Error writing response: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
		_, err := w.Write([]byte("#Error: name is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
	}

	body := r.URL.Query().Get("body")
	if body == "" {
		_, err := w.Write([]byte("#Error: body is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	err := a.store.TemplateAdd(name, body)
	if err != nil {
		_, err := w.Write([]byte("#Error: " + err.Error()))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
	}

	_, err = w.Write([]byte(fmt.Sprintf("Template %s added", name)))
	if err != nil {
		logger.Error(fmt.Sprintf("Error writing response: %v", err))
	}
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
		_, err := w.Write([]byte("#Error: name is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	b, err := a.store.TemplateGet(name)
	if err != nil {
		_, err := w.Write([]byte("#Error: " + err.Error()))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}
	_, err = w.Write([]byte(b))
	if err != nil {
		logger.Error(fmt.Sprintf("Error writing response: %v", err))
	}
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
		_, err := w.Write([]byte("#Error: name is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	body := r.URL.Query().Get("body")
	if body == "" {
		_, err := w.Write([]byte("#Error: body is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	err := a.store.TemplateSet(name, body)
	if err != nil {
		_, err = w.Write([]byte("#Error: " + err.Error()))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	_, err = w.Write([]byte(fmt.Sprintf("Template %s edited", name)))
	if err != nil {
		logger.Error(fmt.Sprintf("Error writing response: %v", err))
	}
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
		_, err := w.Write([]byte("#Error: name is empty"))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	err := a.store.TemplateDel(name)
	if err != nil {
		_, err := w.Write([]byte("#Error: " + err.Error()))
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
		return
	}

	_, err = w.Write([]byte(fmt.Sprintf("Template %s deleted", name)))
	if err != nil {
		logger.Error(fmt.Sprintf("Error writing response: %v", err))
	}
}

// @Summary Выполнить шаблон
// @Description Выполняет шаблон
// @Tags Template
// @Produce plain/text
// @Success 200 {array} string
// @Router /templ/exec [get]
// @Param name query string true "Имя шаблона"
// @Param db query string false "Имя базы данных"
// @Param format query string false "Формат вывода (text - по умолчанию, json, raw)"
// @Param args query array false "Список аргументов"
// @x-try-it-out-enabled false
func (a *App) handleTemplateExec(w http.ResponseWriter, r *http.Request) {
	logger.Trace("executing template")
	writer := []byte("#Error: unknown error")
	defer func() {
		_, err := w.Write(writer)
		if err != nil {
			logger.Error(fmt.Sprintf("Error writing response: %v", err))
		}
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

	fmtr, err := format.New(formatStr)
	if err != nil {
		writer = []byte("#Error: " + err.Error())
		return
	}
	writer = fmtr.Process(b)
}
