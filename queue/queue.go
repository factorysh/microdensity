package queue

import (
	"fmt"
	"sync"

	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/task"
	"github.com/oleiade/lane"
)

// Queue struct use to put and get job items
type Queue struct {
	sync.RWMutex
	items      *lane.Queue
	runner     *run.Runner
	BatchEnded chan bool
	working    bool
}

// NewQueue inits a new queue struct
func NewQueue(s *Storage, runner *run.Runner) Queue {
	return Queue{
		items:      lane.NewQueue(),
		BatchEnded: make(chan bool, 1),
		runner:     runner,
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
		q.working = true
		go q.DequeueWhile()
	}
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
	workers := make(chan int, maxDequeue)

	for q.items.Head() != nil {
		workers <- 1

		item := q.dequeue()

		t, ok := item.(*task.Task)
		if !ok {
			fmt.Println("error when casting item to task")
			continue
		}

		go func(t *task.Task) {
			_, err := q.runner.Run(t)
			if err != nil {
				fmt.Println(err)
			}
			<-workers
		}(t)
	}

	q.BatchEnded <- true

	q.working = false
}
