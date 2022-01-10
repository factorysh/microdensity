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
	lock    *sync.Mutex
	store   *bbolt.DB
	encoder *gob.Encoder
	decoder *gob.Decoder
	buffer  bytes.Buffer
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
		lock:    &sync.Mutex{},
		encoder: gob.NewEncoder(&b),
		decoder: gob.NewDecoder(&b),
		buffer:  b,
		store:   store,
	}, nil
}

func (q *Queue) Set(t *task.Task) error {
	var err error
	if t.Id == uuid.Nil {
		t.Id, err = uuid.NewUUID()
		if err != nil {
			return err
		}
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	err = q.encoder.Encode(t)
	if err != nil {
		return err
	}
	err = q.encoder.Encode(t)
	if err != nil {
		return err
	}
	if err := q.store.Update(func(tx *bbolt.Tx) error {
		tx.Bucket([]byte(queue)).Put(t.Id[:], q.buffer.Bytes())
		return nil
	}); err != nil {
		return err
	}

	return err
}

func (q *Queue) Get(id uuid.UUID) (*task.Task, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	var t task.Task
	if err := q.store.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(queue)).Get(id.NodeID())
		_, err := q.buffer.Write(v)
		if err != nil {
			return err
		}
		err = q.decoder.Decode(t)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &t, nil
}

func (q *Queue) Length() (int, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	var size int
	if err := q.store.View(func(tx *bbolt.Tx) error {
		size = tx.Bucket([]byte(queue)).Stats().KeyN
		return nil
	}); err != nil {
		return 0, err
	}
	return size, nil
}
