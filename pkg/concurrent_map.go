package pkg

import (
	"reflect"
	"sync"
)

type ConcurrentMap[K comparable, V any] struct {
	m *sync.Map
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		m: &sync.Map{},
	}
}

func (m *ConcurrentMap[K, V]) Load(key K) (V, bool) {
	var zero V

	anyVal, ok := m.m.Load(key)
	if !ok {
		return zero, false
	}

	value, isType := anyVal.(V)
	if !isType {
		panic("wrong type value is saved: " + reflect.TypeOf(anyVal).String())
	}

	return value, true
}

func (m *ConcurrentMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *ConcurrentMap[K, V]) Clear() {
	m.m.Clear()
}

func (m *ConcurrentMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	var zero V

	actualAny, loaded := m.m.LoadOrStore(key, value)
	if !loaded {
		return zero, false
	}

	actual, isType := actualAny.(V)
	if !isType {
		panic("wrong type value is saved: " + reflect.TypeOf(actualAny).String())
	}
	return actual, true
}

func (m *ConcurrentMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	var zero V

	anyValue, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return zero, false
	}

	value, isType := anyValue.(V)
	if !isType {
		panic("wrong type value is saved: " + reflect.TypeOf(anyValue).String())
	}
	return value, true
}

func (m *ConcurrentMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	var zero V

	previousAny, loaded := m.m.Swap(key, value)
	if !loaded {
		return zero, false
	}

	previous, isType := previousAny.(V)
	if !isType {
		panic("wrong type value is saved: " + reflect.TypeOf(previousAny).String())
	}
	return previous, true
}
