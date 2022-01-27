package queue

import (
	"fmt"
	"sync"

	"github.com/factorysh/microdensity/task"
	"github.com/oleiade/lane"
)

// Queue struct use to put and get job items
type Queue struct {
	sync.RWMutex
	items      *lane.Queue
	BatchEnded chan bool
	working    bool
}

// NewQueue inits a new queue struct
func NewQueue(s *Storage) Queue {
	return Queue{
		items:      lane.NewQueue(),
		BatchEnded: make(chan bool, 1),
	}
}

// Len of the current dequeue
func (q *Queue) Len() int {
	q.RLock()
	defer q.RUnlock()

	return q.items.Size()
}

// Put a new item into the queue and the storage
func (q *Queue) Put(item *task.Task) {
	q.Lock()
	defer q.Unlock()

	q.items.Enqueue(item)

	if !q.working {
		go q.DequeueWhile()
	}
}

// toogleWorking is used to toggle the working var
func (q *Queue) toogleWorking() {
	q.Lock()
	defer q.Unlock()

	q.working = !q.working
}

// dequeue one item
func (q *Queue) dequeue() interface{} {
	q.Lock()
	defer q.Unlock()

	return q.items.Dequeue()
}

const maxDequeue = 1

// DequeueWhile start maxDequeue workers while the queue is not empty
func (q *Queue) DequeueWhile() {
	q.toogleWorking()

	workers := make(chan int, maxDequeue)

	for q.items.Head() != nil {
		workers <- 1

		item := q.dequeue()

		go func(interface{}) {
			fmt.Println("WIP", item)
			<-workers
		}(item)
	}

	q.BatchEnded <- true

	q.toogleWorking()
}
