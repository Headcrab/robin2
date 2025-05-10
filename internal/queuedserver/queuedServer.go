package queuedserver

import (
	"net/http"
)

type QueuedServer struct {
	server       *http.Server
	requestQueue chan *http.Request
	workerPool   chan struct{}
}

func NewQueuedServer(addr string, handler http.Handler, queueSize, workerCount int) *QueuedServer {
	return &QueuedServer{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		requestQueue: make(chan *http.Request, queueSize),
		workerPool:   make(chan struct{}, workerCount),
	}
}

func (qs *QueuedServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case qs.requestQueue <- r:
		// Запрос успешно добавлен в очередь
	default:
		// Очередь переполнена, отправляем ошибку клиенту
		http.Error(w, "Server is too busy, please try again later", http.StatusServiceUnavailable)
	}
}

func (qs *QueuedServer) processRequests() {
	for r := range qs.requestQueue {
		qs.workerPool <- struct{}{}
		go func(r *http.Request) {
			defer func() { <-qs.workerPool }()
			qs.server.Handler.ServeHTTP(newResponseWriter(r), r)
		}(r)
	}
}

func (qs *QueuedServer) ListenAndServe() error {
	go qs.processRequests()

	qs.server.Handler = http.HandlerFunc(qs.ServeHTTP)
	return qs.server.ListenAndServe()
}

type responseWriter struct {
	http.ResponseWriter
	request *http.Request
}

func newResponseWriter(r *http.Request) *responseWriter {
	return &responseWriter{
		ResponseWriter: nil,
		request:        r,
	}
}

func (rw *responseWriter) Header() http.Header {
	return make(http.Header)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {}

// Использование:
// func main() {
// 	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

// 	queuedServer := NewQueuedServer(":8080", yourHandler, 1000, 100, logger)

// 	if err := queuedServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			logger.Fatal(err.Error())
// 	}
// }
