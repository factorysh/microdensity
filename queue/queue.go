package queue

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

const queue = "queue"

type Queue struct {
	store    *bbolt.DB
	encoding *Encoding
}

func New(store *bbolt.DB) (*Queue, error) {

	if err := store.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(queue))
		return err
	}); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	return &Queue{
		encoding: &Encoding{
			lock:    &sync.Mutex{},
			encoder: gob.NewEncoder(&b),
			decoder: gob.NewDecoder(&b),
			buffer:  b,
		},
		store: store,
	}, nil
}

func (q *Queue) Put(t *task.Task) error {
	var err error
	if t.Id == uuid.Nil {
		t.Id, err = uuid.NewUUID() // it's v1 UUID, with a timestamp
		if err != nil {
			return err
		}
	}
	raw, err := q.encoding.Encode(t)
	if err != nil {
		return err
	}
	if err := q.store.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(queue)).Put(t.Id[:], raw)
	}); err != nil {
		return err
	}

	return err
}

func (q *Queue) Get(id uuid.UUID) (*task.Task, error) {
	var t *task.Task
	if err := q.store.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(queue)).Get(id.NodeID())
		var err error
		t, err = q.encoding.Decode(v)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return t, nil
}

func (q *Queue) Length() (int, error) {
	var size int
	if err := q.store.View(func(tx *bbolt.Tx) error {
		size = tx.Bucket([]byte(queue)).Stats().KeyN
		return nil
	}); err != nil {
		return 0, err
	}
	return size, nil
}

func (q *Queue) First(state task.State) (*task.Task, error) {
	var t *task.Task

	if err := q.store.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte(queue)).Cursor()
		var err2 error
		for k, v := c.First(); k != nil; k, v = c.Next() {
			t, err2 = q.encoding.Decode(v)
			if err2 != nil {
				return err2
			}
			if t.State == state {
				break
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return t, nil
}
