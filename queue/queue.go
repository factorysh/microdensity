package queue

import (
	"fmt"
	"sync"

	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/oleiade/lane"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	queueAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "microdensity_queue_task_added_total",
		Help: "The total number task added",
	})

	queueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "microdensity_queue_size",
		Help: "Queue size",
	})
)

// Queue struct use to put and get job items
type Queue struct {
	sync.RWMutex
	items      *lane.Queue
	runner     *run.Runner
	storage    storage.Storage
	BatchEnded chan bool
	working    bool
}

// NewQueue inits a new queue struct
func NewQueue(s storage.Storage, runner *run.Runner) Queue {
	queueSize.Set(0)
	return Queue{
		items:      lane.NewQueue(),
		BatchEnded: make(chan bool, 1),
		runner:     runner,
		storage:    s,
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

	queueAdded.Inc()
	queueSize.Inc()

	if !q.working {
		q.working = true
		go q.DequeueWhile()
	}
}

// dequeue one item
func (q *Queue) dequeue() interface{} {
	q.Lock()
	defer q.Unlock()

	queueSize.Dec()
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
			defer func() {
				<-workers
			}()

			ret := -1

			t.State = task.Running
			err := q.storage.Upsert(t)
			// FIXME: handle err
			if err != nil {
				fmt.Println(err)
				return
			}

			ret, err = q.runner.Run(t)
			if err != nil {
				fmt.Println(err)
			}

			if ret == 0 {
				t.State = task.Done
			} else {
				t.State = task.Failed
			}

			err = q.storage.Upsert(t)
			// FIXME: handle err
			if err != nil {
				fmt.Println(err)
				return
			}
		}(t)
	}

	q.BatchEnded <- true

	q.working = false
}
