package pool

import (
	"container/heap"
)

// PriorityQueue
type PriorityQueue struct {
	q queue
	// lock sync.RWMutex
}

// NewPriorityQueue return a priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{q: make(queue, 0)} // lock: sync.RWMutex{},

	heap.Init(&pq.q)
	return pq
}

// Push a value with priority, 0 is highest priority
func (pq *PriorityQueue) Push(v interface{}, priority int64) {
	// pq.lock.Lock()
	// defer pq.lock.Unlock()

	heap.Push(&pq.q, &node{priority: priority, value: v})
}

// Pop a data and remove from queue
func (pq *PriorityQueue) Pop() interface{} {
	// pq.lock.Lock()
	// defer pq.lock.Unlock()

	if pq.q.Len() == 0 {
		return nil
	}

	return heap.Pop(&pq.q).(*node).value
}

// Size return the element count
func (pq *PriorityQueue) Size() int {
	// pq.lock.RLock()
	// defer pq.lock.RUnlock()

	return pq.q.Len()
}

//
// memory cache
//

type node struct {
	value    interface{}
	priority int64
}

type queue []*node

func (q queue) Len() int           { return len(q) }
func (q queue) Less(i, j int) bool { return q[i].priority > q[j].priority }
func (q queue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *queue) Push(x interface{}) { *q = append(*q, x.(*node)) }
func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}
