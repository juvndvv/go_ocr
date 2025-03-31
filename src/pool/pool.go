package pool

import (
	"context"
	"errors"
	"sync"
	"time"
)

type TaskID string
type TaskFunc func() (interface{}, error)

type Task struct {
	ID     TaskID
	Action TaskFunc
}

type Result struct {
	TaskID TaskID
	Data   interface{}
	Error  error
}

type WorkerPool struct {
	tasks         chan Task
	results       chan Result
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	config        Config
	activeWorkers int
	mu            sync.Mutex
}

type Config struct {
	MaxWorkers    int           // Máximo de goroutines paralelas
	MaxTasks      int           // Tamaño del buffer de tareas
	Timeout       time.Duration // Timeout por tarea
	Retries       int           // Número de reintentos por tarea
	ResultBufSize int           // Tamaño del buffer de resultados
}

type Option func(*Config)

func NewWorkerPool(opts ...Option) *WorkerPool {
	// Configuración por defecto
	config := Config{
		MaxWorkers:    5,
		MaxTasks:      100,
		Timeout:       30 * time.Second,
		Retries:       3,
		ResultBufSize: 50,
	}

	// Aplicar opciones
	for _, opt := range opts {
		opt(&config)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		tasks:   make(chan Task, config.MaxTasks),
		results: make(chan Result, config.ResultBufSize),
		ctx:     ctx,
		cancel:  cancel,
		config:  config,
	}
}

func WithMaxWorkers(n int) Option {
	return func(c *Config) {
		c.MaxWorkers = n
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.Timeout = d
	}
}

func WithRetries(n int) Option {
	return func(c *Config) {
		c.Retries = n
	}
}

func WithTaskBufferSize(n int) Option {
	return func(c *Config) {
		c.MaxTasks = n
	}
}

func (wp *WorkerPool) Start() {
	wp.wg.Add(wp.config.MaxWorkers)
	for i := 0; i < wp.config.MaxWorkers; i++ {
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			wp.processTask(task)
		}
	}
}

func (wp *WorkerPool) processTask(task Task) {
	var result Result
	var err error
	var data interface{}

	ctx, cancel := context.WithTimeout(wp.ctx, wp.config.Timeout)
	defer cancel()

	for attempt := 0; attempt <= wp.config.Retries; attempt++ {
		select {
		case <-ctx.Done():
			err = errors.New("timeout excedido")
			break
		default:
			data, err = task.Action()
			if err == nil {
				break
			}
		}
	}

	result = Result{
		TaskID: task.ID,
		Data:   data,
		Error:  err,
	}

	select {
	case wp.results <- result:
	case <-wp.ctx.Done():
	}
}

func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		return errors.New("worker pool detenido")
	default:
		return errors.New("buffer de tareas lleno")
	}
}

func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.tasks)
	wp.wg.Wait()
	close(wp.results)
}

func (wp *WorkerPool) ActiveWorkers() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.activeWorkers
}

func (wp *WorkerPool) PendingTasks() int {
	return len(wp.tasks)
}
