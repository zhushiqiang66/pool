package pool

import (
	"time"
)

// Task is function
type Task func()

// Worker is the plan for task, one task have one worker
type Worker struct {
	pool    *Pool
	task    chan Task
	recycle time.Time

	stop chan struct{}
}

func NewWorker(p *Pool) *Worker {
	w := &Worker{
		pool:    p,
		task:    make(chan Task),
		stop:    make(chan struct{}),
		recycle: time.Now(),
	}

	go w.run()
	return w
}

// run
func (w *Worker) run() {
	for {
		select {
		case <-w.stop:
			return
		case f := <-w.task:
			f()
			w.pool.RevertWorker(w)
		}
	}
}

// Stop
func (w *Worker) Stop() {
	w.stop <- struct{}{}
}

// push
func (w *Worker) push(task Task) {
	w.task <- task
}
