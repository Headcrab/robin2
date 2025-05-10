package pool

import (
	"log"
	"sync"
	"time"
)

// Task представляет функцию, которую нужно выполнить
type Task func() []byte

// WorkerPool управляет пулом воркеров
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	resultQueue chan []byte
	wg          sync.WaitGroup
	logger      *log.Logger
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(workerCount int, logger *log.Logger) *WorkerPool {
	wp := &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task),
		resultQueue: make(chan []byte, workerCount),
		logger:      logger,
	}

	if wp.logger != nil {
		wp.logger.Printf("Creating new worker pool with %d workers", workerCount)
	}
	wp.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go wp.worker(i)
	}

	return wp
}

// worker выполняет задачи из очереди
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	if wp.logger != nil {
		wp.logger.Printf("Worker %d started", id)
	}
	for task := range wp.taskQueue {
		if wp.logger != nil {
			wp.logger.Printf("Worker %d started task", id)
		}
		startTime := time.Now()
		result := task()
		duration := time.Since(startTime)
		if wp.logger != nil {
			wp.logger.Printf("Worker %d completed task in %v", id, duration)
		}
		wp.resultQueue <- result
	}
	if wp.logger != nil {
		wp.logger.Printf("Worker %d stopped", id)
	}
}

// Submit отправляет задачу в очередь
func (wp *WorkerPool) Submit(task Task) {
	if wp.logger != nil {
		wp.logger.Printf("Task submitted to queue")
	}
	wp.taskQueue <- task
}

// GetResult получает результат выполнения задачи
func (wp *WorkerPool) GetResult() []byte {
	result := <-wp.resultQueue
	if wp.logger != nil {
		wp.logger.Printf("Result retrieved from queue")
	}
	return result
}

// Close закрывает пул воркеров и ожидает завершения всех задач
func (wp *WorkerPool) Close() {
	if wp.logger != nil {
		wp.logger.Printf("Closing worker pool")
	}
	close(wp.taskQueue)
	wp.wg.Wait()
	close(wp.resultQueue)
	if wp.logger != nil {
		wp.logger.Printf("Worker pool closed")
	}
}

// ProcessQueued выполняет задачу в пуле и возвращает результат
func (wp *WorkerPool) ProcessQueued(task Task) []byte {
	if wp.logger != nil {
		wp.logger.Printf("Processing queued task")
	}
	wp.Submit(task)
	return wp.GetResult()
}
