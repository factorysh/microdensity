package queue

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

const queue = "queue"

type Storage struct {
	store    *bbolt.DB
	encoding *Encoding
	logger   *zap.Logger
}

func New(store *bbolt.DB) (*Storage, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	if err := store.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(queue))
		if err != nil {
			logger.Error("Bbolt can't create bucket", zap.String("bucket", queue), zap.Error(err))
		}
		return err
	}); err != nil {
		logger.Error("BBolt store error", zap.Error(err))
		return nil, err
	}
	var b bytes.Buffer
	return &Storage{
		encoding: &Encoding{
			lock:    &sync.Mutex{},
			encoder: gob.NewEncoder(&b),
			decoder: gob.NewDecoder(&b),
			buffer:  b,
		},
		store:  store,
		logger: logger,
	}, nil
}

func (q *Storage) Put(t *task.Task) error {
	l := q.logger.With(zap.Any("task", t))
	var err error
	if t.Id == uuid.Nil {
		t.Id, err = uuid.NewUUID() // it's v1 UUID, with a timestamp
		if err != nil {
			l.Error("UUID error", zap.Error(err))
			return err
		}
	}
	raw, err := q.encoding.Encode(t)
	if err != nil {
		l.Error("encoding error", zap.Error(err))
		return err
	}
	if err := q.store.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(queue)).Put(t.Id[:], raw)
	}); err != nil {
		l.Error("bbolt update error", zap.Error(err))
		return err
	}

	return err
}

func (q *Storage) Get(id uuid.UUID) (*task.Task, error) {
	l := q.logger.With(zap.String("uuid", id.String()))
	var t *task.Task
	if err := q.store.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(queue)).Get(id.NodeID())
		var err error
		t, err = q.encoding.Decode(v)
		if err != nil {
			l.Error("Encoding error", zap.Error(err))
			return err
		}
		return nil
	}); err != nil {
		l.Error("bbolt Get error", zap.Error(err))
		return nil, err
	}
	return t, nil
}

func (q *Storage) Length() (int, error) {
	var size int
	if err := q.store.View(func(tx *bbolt.Tx) error {
		size = tx.Bucket([]byte(queue)).Stats().KeyN
		return nil
	}); err != nil {
		q.logger.Error("Length error", zap.Error(err))
		return 0, err
	}
	return size, nil
}

func (q *Storage) First(state task.State) (*task.Task, error) {
	var t *task.Task
	l := q.logger.With(zap.String("state", state.String()))

	if err := q.store.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte(queue)).Cursor()
		var err2 error
		for k, v := c.First(); k != nil; k, v = c.Next() {
			t, err2 = q.encoding.Decode(v)
			if err2 != nil {
				l.Error("decode error", zap.Error(err2))
				return err2
			}
			if t.State == state {
				break
			}
		}
		return nil
	}); err != nil {
		l.Error("first error", zap.Error(err))
		return nil, err
	}
	return t, nil
}
