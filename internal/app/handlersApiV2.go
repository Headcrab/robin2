package robin

import (
	"net/http"
	"strings"
)

func (a *App) handleAPIV2GetTagOnDate(w http.ResponseWriter, r *http.Request) {
	writer := []byte("Error: unknown error")
	defer func() {
		w.Write(writer)
	}()
	a.opCount++

	// Extract path parameters
	basePath := "/api/v2/get/"
	params := strings.Split(r.URL.Path[len(basePath):], "/")
	writer = []byte(strings.Join(params, "\n"))

	// query := r.URL.Query()
	// tag := query.Get("tag")
	// date := query.Get("date")
	// from := query.Get("from")
	// to := query.Get("to")
	// group := query.Get("group")
	// roundStr := query.Get("round")
	// count := query.Get("count")
	// format := query.Get("format")

	// round := utils.ThenIf(roundStr != "", func() int { r, _ := strconv.Atoi(roundStr); return r }(), a.config2.Round)

}
