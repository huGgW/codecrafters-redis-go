package event

import "io"

const (
	ReadEventType  Type = "read"
	WriteEventType Type = "write"
	CloseEventType Type = "close"
)

type ReadEvent struct {
	Reader io.Reader
}

func (r *ReadEvent) Type() Type {
	return ReadEventType
}

type WriteEvent struct {
	Data   []byte
	Writer io.Writer
}

func (w *WriteEvent) Type() Type {
	return WriteEventType
}

type CloseEvent struct {
	Closer io.Closer
}

func (c *CloseEvent) Type() Type {
	return CloseEventType
}
