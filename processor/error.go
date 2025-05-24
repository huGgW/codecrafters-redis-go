package processor

import (
	"log/slog"

	"github.com/codecrafters-io/redis-starter-go/event"
)

type ErrorHandler struct {
	logger *slog.Logger
}

func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

func (h *ErrorHandler) Target() event.Type {
	return event.ErrorEventType
}

func (h *ErrorHandler) Handle(e event.Event, _ func(event.Event)) error {
	if e.Type() != event.ErrorEventType {
		return event.ErrInvalidEventType
	}

	errorEvent, ok := e.(*event.ErrorEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	h.logger.Error("error occurred while processing event", slog.Any("event", errorEvent.Event), slog.Any("error", errorEvent.Err))
	return nil
}
