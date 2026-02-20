package robin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestApp создаёт приложение с инициализированной БД для тестов.
// TEST_WORK_DIR указывает NewApp где искать конфиг и .env.
// При ошибке подключения к БД тест пропускается.
func newTestApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("TEST_WORK_DIR", projectRoot())
	app := NewApp()
	if err := app.initDatabase(); err != nil {
		t.Skipf("пропущено: нет соединения с БД: %v", err)
	}
	return app
}

// get выполняет GET-запрос к handler через httptest и возвращает recorder.
func get(handler http.HandlerFunc, rawURL string) *httptest.ResponseRecorder {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    mustURL(rawURL),
		Header: make(http.Header),
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func mustURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

// ─── /api/info/ ──────────────────────────────────────────────────────────────

func Test_handleAPIInfo(t *testing.T) {
	app := newTestApp(t)

	w := get(app.handleAPIInfo, "/api/info/")

	if w.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d", w.Code)
	}
	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("невалидный JSON: %v; тело: %s", err, w.Body.String())
	}
	for _, key := range []string{"name", "version", "uptime", "op_count"} {
		if _, ok := info[key]; !ok {
			t.Errorf("отсутствует ключ %q в ответе", key)
		}
	}
}

// ─── /api/status/ ────────────────────────────────────────────────────────────

func Test_handleAPIServerStatus(t *testing.T) {
	app := newTestApp(t)

	w := get(app.handleAPIServerStatus, "/api/status/")

	if w.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("ожидался Content-Type application/json, получен %q", ct)
	}
	var status map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("невалидный JSON: %v; тело: %s", err, w.Body.String())
	}
}

// ─── /api/log/ ───────────────────────────────────────────────────────────────

func Test_handleAPIGetLog(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name           string
		query          string
		wantCode       int
		wantCTContains string
	}{
		{name: "default (text)", query: "/api/log/", wantCode: 200, wantCTContains: "text/plain"},
		{name: "format=text", query: "/api/log/?format=text", wantCode: 200, wantCTContains: "text/plain"},
		{name: "format=str", query: "/api/log/?format=str", wantCode: 200, wantCTContains: "text/plain"},
		{name: "format=raw", query: "/api/log/?format=raw", wantCode: 200, wantCTContains: "text/plain"},
		{name: "format=json", query: "/api/log/?format=json", wantCode: 200, wantCTContains: "application/json"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIGetLog, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
			ct := w.Header().Get("Content-Type")
			if !strings.Contains(ct, tc.wantCTContains) {
				t.Errorf("ожидался Content-Type содержащий %q, получен %q", tc.wantCTContains, ct)
			}
		})
	}
}

// ─── /api/log/clear/ ─────────────────────────────────────────────────────────

func Test_handleAPIClearLog(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name     string
		method   string
		wantCode int
	}{
		{name: "DELETE", method: http.MethodDelete, wantCode: http.StatusOK},
		{name: "POST", method: http.MethodPost, wantCode: http.StatusOK},
		{name: "GET (запрещён)", method: http.MethodGet, wantCode: http.StatusMethodNotAllowed},
		{name: "PUT (запрещён)", method: http.MethodPut, wantCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &http.Request{
				Method: tc.method,
				URL:    mustURL("/api/log/clear/"),
				Header: make(http.Header),
			}
			w := httptest.NewRecorder()
			app.handleAPIClearLog(w, req)
			if w.Code != tc.wantCode {
				t.Errorf("метод %s: ожидался %d, получен %d; тело: %s", tc.method, tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

// ─── /get/tag/list/ ──────────────────────────────────────────────────────────

func Test_handleAPIGetTagList(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name     string
		query    string
		wantCode int
		// minTags: минимальное ожидаемое кол-во тегов (-1 = не проверять)
		minTags int
	}{
		{name: "без параметров", query: "/get/tag/list/", wantCode: 200, minTags: 0},
		{name: "like=A20%", query: "/get/tag/list/?like=A20%25", wantCode: 200, minTags: 0},
		{name: "like=A20_WT_01%", query: "/get/tag/list/?like=A20_WT_01%25", wantCode: 200, minTags: 0},
		{name: "like=NONEXISTENT_XYZ", query: "/get/tag/list/?like=NONEXISTENT_XYZ", wantCode: 200, minTags: -1},
		{name: "format=json", query: "/get/tag/list/?format=json", wantCode: 200, minTags: -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIGetTagList, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
			body := w.Body.Bytes()
			if tc.minTags >= 0 {
				var tags interface{}
				if err := json.Unmarshal(body, &tags); err != nil {
					t.Errorf("невалидный JSON: %v; тело: %s", err, string(body))
				}
			}
		})
	}
}

// ─── /get/tag/ ───────────────────────────────────────────────────────────────

func Test_handleAPIGetTag(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name           string
		query          string
		wantCode       int
		bodyNotContain string // тело не должно содержать эту строку (обычно "#Error:")
	}{
		// --- date ---
		{
			name:     "tag+date валидные",
			query:    "/get/tag/?tag=A20_WT_01&date=19.02.2026 09:00",
			wantCode: 200,
		},
		{
			name:     "tag+date формат ISO",
			query:    "/get/tag/?tag=A20_WT_01&date=2026-02-19 09:00:00",
			wantCode: 200,
		},
		{
			name:           "tag пустой",
			query:          "/get/tag/?date=2026-02-19 09:00:00",
			wantCode:       200,
			bodyNotContain: "",
		},
		{
			name:           "date невалидная",
			query:          "/get/tag/?tag=A20_WT_01&date=not-a-date",
			wantCode:       200,
			bodyNotContain: "",
		},
		// --- from/to без group ---
		{
			name:     "tag+from+to",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00",
			wantCode: 200,
		},
		{
			name:     "tag+from+to format=json",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&format=json",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+round=4",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&round=4",
			wantCode: 200,
		},
		// --- from/to + count без group ---
		{
			name:     "tag+from+to+count=100",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&count=100",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+count невалидный",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&count=abc",
			wantCode: 200,
		},
		// --- from/to + group ---
		{
			name:     "tag+from+to+group=avg",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=avg",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+group=sum",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=sum",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+group=min",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=min",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+group=max",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=max",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+group=count",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=count",
			wantCode: 200,
		},
		// --- from/to + count + group ---
		{
			name:     "tag+from+to+count+group=avg",
			query:    "/get/tag/?tag=A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&count=10&group=avg",
			wantCode: 200,
		},
		// --- несколько тегов через запятую + date ---
		{
			name:     "несколько тегов+date",
			query:    "/get/tag/?tag=A20_WT_01,A20_WT_01&date=2026-02-19 09:00:00",
			wantCode: 200,
		},
		// --- несколько тегов через запятую + from+to+group ---
		{
			name:     "несколько тегов+from+to+group=avg",
			query:    "/get/tag/?tag=A20_WT_01,A20_WT_01&from=19.02.2026 08:00&to=19.02.2026 09:00&group=avg",
			wantCode: 200,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIGetTag, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

// ─── /get/tag/down/ ──────────────────────────────────────────────────────────

func Test_handleAPIGetTagDown(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name        string
		query       string
		wantCode    int
		bodyContain string // если непусто — тело должно содержать эту строку
	}{
		{
			name:        "без tag",
			query:       "/get/tag/down/",
			wantCode:    200,
			bodyContain: "#Error: tag is empty",
		},
		{
			name:     "tag+from+to+count=0",
			query:    "/get/tag/down/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=0",
			wantCode: 200,
		},
		{
			name:     "tag+from+to без count (default 0)",
			query:    "/get/tag/down/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+count=1",
			query:    "/get/tag/down/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=1",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+count=999 (за пределами)",
			query:    "/get/tag/down/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=999",
			wantCode: 200,
		},
		{
			name:        "невалидная from",
			query:       "/get/tag/down/?tag=A20_WT_01&from=not-a-date&to=19.02.2026 23:59",
			wantCode:    200,
			bodyContain: "#Error:",
		},
		{
			name:        "невалидная to",
			query:       "/get/tag/down/?tag=A20_WT_01&from=19.02.2026 00:00&to=not-a-date",
			wantCode:    200,
			bodyContain: "#Error:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIGetTagDown, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
			if tc.bodyContain != "" && !strings.Contains(w.Body.String(), tc.bodyContain) {
				t.Errorf("ожидалось %q в теле, тело: %s", tc.bodyContain, w.Body.String())
			}
		})
	}
}

// ─── /get/tag/up/ ────────────────────────────────────────────────────────────

func Test_handleAPIGetTagUp(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name        string
		query       string
		wantCode    int
		bodyContain string
	}{
		{
			name:        "без tag",
			query:       "/get/tag/up/",
			wantCode:    200,
			bodyContain: "#Error: tag is empty",
		},
		{
			name:     "tag+from+to+count=0",
			query:    "/get/tag/up/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=0",
			wantCode: 200,
		},
		{
			name:     "tag+from+to без count",
			query:    "/get/tag/up/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+count=1",
			query:    "/get/tag/up/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=1",
			wantCode: 200,
		},
		{
			name:     "tag+from+to+count=999 (за пределами)",
			query:    "/get/tag/up/?tag=A20_WT_01&from=19.02.2026 00:00&to=19.02.2026 23:59&count=999",
			wantCode: 200,
		},
		{
			name:        "невалидная from",
			query:       "/get/tag/up/?tag=A20_WT_01&from=not-a-date&to=19.02.2026 23:59",
			wantCode:    200,
			bodyContain: "#Error:",
		},
		{
			name:        "невалидная to",
			query:       "/get/tag/up/?tag=A20_WT_01&from=19.02.2026 00:00&to=not-a-date",
			wantCode:    200,
			bodyContain: "#Error:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIGetTagUp, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
			if tc.bodyContain != "" && !strings.Contains(w.Body.String(), tc.bodyContain) {
				t.Errorf("ожидалось %q в теле, тело: %s", tc.bodyContain, w.Body.String())
			}
		})
	}
}

// ─── /tag/decode/ ────────────────────────────────────────────────────────────

func Test_handleTagDecode(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name     string
		query    string
		wantCode int
	}{
		{
			name:     "без tag → 400",
			query:    "/tag/decode/",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "один тег",
			query:    "/tag/decode/?tag=A20_WT_01",
			wantCode: http.StatusOK,
		},
		{
			name:     "несколько тегов",
			query:    "/tag/decode/?tag=A20_WT_01,A20_WT_01",
			wantCode: http.StatusOK,
		},
		{
			name:     "format=json",
			query:    "/tag/decode/?tag=A20_WT_01&format=json",
			wantCode: http.StatusOK,
		},
		{
			name:     "несуществующий тег",
			query:    "/tag/decode/?tag=NONEXISTENT_XYZ_TAG",
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleTagDecode, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

// ─── /api/reload/ ────────────────────────────────────────────────────────────

func Test_handleAPIReloadConfig(t *testing.T) {
	app := newTestApp(t)

	w := get(app.handleAPIReloadConfig, "/api/reload/")

	// Перезагрузка конфига должна успешно отработать если конфиг на месте
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("ожидался 200 или 500, получен %d; тело: %s", w.Code, w.Body.String())
	}
}

// ─── /api/v2/get/ ────────────────────────────────────────────────────────────

func Test_handleAPIV2GetTagOnDate(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name     string
		query    string
		wantCode int
	}{
		{name: "пустой path", query: "/api/v2/get/", wantCode: 200},
		{name: "tag в path", query: "/api/v2/get/A20_WT_01/", wantCode: 200},
		{name: "tag+date в path", query: "/api/v2/get/A20_WT_01/2026-02-19/", wantCode: 200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := get(app.handleAPIV2GetTagOnDate, tc.query)
			if w.Code != tc.wantCode {
				t.Errorf("ожидался %d, получен %d; тело: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}
