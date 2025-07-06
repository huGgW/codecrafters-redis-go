package pkg

import libHeap "container/heap"

type Heap[T comparable] struct {
	heap *heap[T]
}

func NewHeap[T comparable](less func(x, y T) bool) *Heap[T] {
	h := &heap[T]{
		data: []T{},
		less: less,
	}

	libHeap.Init(h)
	return &Heap[T]{heap: h}
}

func (h *Heap[T]) Push(x T) {
	libHeap.Push(h.heap, x)
}

func (h *Heap[T]) Pop() (T, bool) {
	var zero T

	if h.Len() == 0 {
		return zero, false
	}

	val := libHeap.Pop(h.heap)
	return val.(T), true
}

func (h *Heap[T]) Len() int {
	return h.heap.Len()
}

func (h *Heap[T]) Peek() (T, bool) {
	var zero T

	if h.Len() == 0 {
		return zero, false
	}

	val := h.heap.data[0]
	return val, true
}

var _ libHeap.Interface = (*heap[int])(nil)

type heap[T comparable] struct {
	data []T
	less func(x, y T) bool
}

func (h *heap[T]) Len() int {
	return len(h.data)
}

func (h *heap[T]) Less(i int, j int) bool {
	x := h.data[i]
	y := h.data[j]

	return h.less(x, y)
}

func (h *heap[T]) Pop() any {
	if len(h.data) == 0 {
		return nil
	}

	val := h.data[0]
	h.data = h.data[1:]

	return val
}

func (h *heap[T]) Push(x any) {
	h.data = append(h.data, x.(T))
}

func (h *heap[T]) Swap(i int, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}
