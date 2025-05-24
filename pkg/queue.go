package queue

import (
	"container/list"
	"sync"
)

type ConcurrentQueue[T any] struct {
	list *list.List
	lock sync.RWMutex
}

func NewConcurrentQueue[T any]() *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{
		list: list.New(),
		lock: sync.RWMutex{},
	}
}

func (q *ConcurrentQueue[T]) IsEmpty() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return q.list.Len() == 0
}

func (q *ConcurrentQueue[T]) Enqueue(val T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	_ = q.list.PushBack(val)
}

func (q *ConcurrentQueue[T]) Dequeue() (T, bool) {
	var zero T

	q.lock.Lock()
	defer q.lock.Unlock()

	if elem := q.list.Front(); elem != nil {
		return q.list.Remove(elem).(T), true
	} else {
		return zero, false
	}
}
