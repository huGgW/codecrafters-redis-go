package event

import (
	"errors"
	"time"

	"github.com/codecrafters-io/redis-starter-go/spec"
)

type Type string

type Event interface {
	ID() uint64
	Type() Type
}

type Handler interface {
	Handle(event Event, push func(Event)) error
	Target() Type
}

type Pusher interface {
	InitPushing(push func(Event))
	ShutdownPushing()
}

// events

const (
	ErrorEventType Type = "error"

	ReadEventType  = "read"
	WriteEventType = "write"
	CloseEventType = "close"

	ExpireEventType = "expire"

	LexingEventType  = "lexing"
	ParseEventType   = "parse"
	ExecuteEventType = "execute"
	FormatEventType  = "format"
)

var ErrInvalidEventType = errors.New("invalid event type")

type ErrorEvent struct {
	Event Event
	Err   error
}

func (e *ErrorEvent) ID() uint64 {
	return e.Event.ID()
}

func (e *ErrorEvent) Type() Type {
	return ErrorEventType
}

type ReadEvent struct {
	ID_ uint64
}

func (r *ReadEvent) Type() Type {
	return ReadEventType
}

func (r *ReadEvent) ID() uint64 {
	return r.ID_
}

type WriteEvent struct {
	ID_  uint64
	Data []byte
}

func (w *WriteEvent) ID() uint64 {
	return w.ID_
}

func (w *WriteEvent) Type() Type {
	return WriteEventType
}

type CloseEvent struct {
	ID_ uint64
}

func (c *CloseEvent) ID() uint64 {
	return c.ID_
}

func (c *CloseEvent) Type() Type {
	return CloseEventType
}

type ExpireEvent struct {
	ID_  uint64
	Time time.Time
}

func (e *ExpireEvent) Type() Type {
	return ExpireEventType
}

func (e *ExpireEvent) ID() uint64 {
	return e.ID_
}

type LexingEvent struct {
	ID_  uint64
	Data []byte
}

func (p *LexingEvent) ID() uint64 {
	return p.ID_
}

func (p *LexingEvent) Type() Type {
	return LexingEventType
}

type ParseEvent struct {
	ID_  uint64
	Data spec.Data
}

func (c *ParseEvent) ID() uint64 {
	return c.ID_
}

func (c *ParseEvent) Type() Type {
	return ParseEventType
}

type ExecuteEvent struct {
	ID_     uint64
	Command spec.Command
}

func (e *ExecuteEvent) ID() uint64 {
	return e.ID_
}

func (e *ExecuteEvent) Type() Type {
	return ExecuteEventType
}

type FormatEvent struct {
	ID_  uint64
	Data spec.Data
}

func (c *FormatEvent) ID() uint64 {
	return c.ID_
}

func (c *FormatEvent) Type() Type {
	return FormatEventType
}
