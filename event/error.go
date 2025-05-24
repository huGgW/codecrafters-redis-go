package event

import "errors"

var ErrInvalidEventType = errors.New("invalid event type")

const ErrorEventType Type = "error"

type ErrorEvent struct {
	Event Event
	Err   error
}

func (e *ErrorEvent) Type() Type {
	return ErrorEventType
}
