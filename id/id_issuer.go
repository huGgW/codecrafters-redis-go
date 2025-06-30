package id

import "sync/atomic"

type IDIssuer[T comparable] interface {
	Issue() T
}

type NumIDIssuer struct {
	next atomic.Uint64
}

func (i *NumIDIssuer) Issue() uint64 {
	return i.next.Add(1) - 1
}
