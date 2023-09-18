package pool

import (
	"sync"
	"sync/atomic"
	"time"
)

// Pool
type Pool struct {
	capacity         int32
	running, waiting int32

	expired time.Duration
	stop    chan struct{}

	lock sync.Locker
	cond *sync.Cond

	status  int32 // 0 close, 1 start
	workers *PriorityQueue
}

//
//
//

const (
	defaultCapacitySize          int32 = 1024
	defaultWorkerExpiredInSecond int32 = 3600

	CloseStatus int32 = 0
	StartStatus int32 = 1
)

// New
func New(opts ...Option) *Pool {
	p := &Pool{
		capacity: defaultCapacitySize,
		expired:  time.Duration(defaultWorkerExpiredInSecond) * time.Second,
		running:  0,
		waiting:  0,
		status:   StartStatus,
		workers:  NewPriorityQueue(),
		stop:     make(chan struct{}),
	}

	p.lock = &sync.Mutex{}
	p.cond = sync.NewCond(p.lock)

	for _, opt := range opts {
		opt(p)
	}

	go p.recovery()
	return p
}

// Submit
func (p *Pool) Submit(task Task) {
	if p.IsClosed() {
		return
	}

	if w := p.pop(); w != nil {
		defer atomic.AddInt32(&p.running, 1)
		w.push(task)
	}
}

// pop returns a available worker to run the tasks.
func (p *Pool) pop() *Worker {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		// pool is closed
		if p.IsClosed() {
			return nil
		}

		// first try to fetch the worker from the queue
		if p.workers.Size() > 0 {
			return p.workers.Pop().(*Worker)
		}

		// if the worker queue is empty and we don't run out of the pool capacity,
		// then just spawn a new worker goroutine.
		if p.workers.Size() == 0 && p.running < p.capacity {
			return NewWorker(p)
		}

		// otherwise, we'll have to keep them blocked and wait for at least one worker to be put back into pool.
		// block and wait for an available worker
		atomic.AddInt32(&p.waiting, 1)
		p.cond.Wait()
		atomic.AddInt32(&p.waiting, -1)
	}
}

// recovery idle worker
func (p *Pool) recovery() {
	ticker := time.NewTicker(p.expired)
	defer ticker.Stop()

	for {
		select {
		case <-p.stop:
			return

		case <-ticker.C:
		}

		// recovery idle worker
		workers := []*Worker{}
		{
			p.lock.Lock()
			for {
				w := p.workers.Pop()
				if w != nil && time.Since(w.(*Worker).recycle) >= p.expired {
					workers = append(workers, w.(*Worker))
					continue
				}

				if w != nil {
					p.workers.Push(w, w.(*Worker).recycle.Unix())
				}

				break
			}
			p.lock.Unlock()
		}

		for _, w := range workers {
			w.Stop()
		}

		// There might be a situation where all workers have been cleaned up (no worker is running),
		// while some invokers still are stuck in p.cond.Wait(), then we need to awake those invokers.
		if p.Waiting() > 0 && (p.Free() > 0 || len(workers) > 0) {
			p.cond.Broadcast()
		}
	}
}

//RevertWorker 回收空闲的 worker 到等待队列中
func (p *Pool) RevertWorker(worker *Worker) {
	if p.IsClosed() {
		return
	}

	defer atomic.AddInt32(&p.running, -1)
	{
		p.lock.Lock()
		defer p.lock.Unlock()

		worker.recycle = time.Now()
		p.workers.Push(worker, worker.recycle.Unix())

		p.cond.Signal()
	}
}

// Close
func (p *Pool) Close() {
	if !atomic.CompareAndSwapInt32(&p.status, StartStatus, CloseStatus) {
		return
	}

	p.stop <- struct{}{}

	// stop all worker
	workers := []*Worker{}
	{
		p.lock.Lock()
		for {
			if w := p.workers.Pop(); w != nil {
				workers = append(workers, w.(*Worker))
				continue
			}

			break
		}
		p.lock.Unlock()
	}

	for _, w := range workers {
		w.Stop()
	}

	p.cond.Broadcast()
}

// IsClosed indicates whether the pool is closed.
func (p *Pool) IsClosed() bool {
	return atomic.LoadInt32(&p.status) == int32(CloseStatus)
}

// Waiting returns the number of tasks which are waiting be executed.
func (p *Pool) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

// Running returns the number of workers currently running.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the number of available goroutines to work, -1 indicates this pool is unlimited.
func (p *Pool) Free() int {
	return int(p.capacity) - p.Running()
}
