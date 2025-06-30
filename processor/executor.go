package processor

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/spec"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) ExecuteHandler() *executeHandler {
	return &executeHandler{
		executor: e,
	}
}

func (e *Executor) Execute(cmd spec.Command) (spec.Data, error) {
	switch cmd.(type) {
	case *spec.PingCommand:
		return &spec.SimpleStringData{S: "PONG"}, nil
	case *spec.EchoCommand:
		echoCmd := cmd.(*spec.EchoCommand)
		return &spec.BulkStringData{
			Len: len(echoCmd.Value),
			S:   echoCmd.Value,
		}, nil
	default:
		return nil, fmt.Errorf("invalid command: %+v", cmd)
	}
}

var _ event.Handler = (*executeHandler)(nil)

type executeHandler struct {
	executor *Executor
}

func (h *executeHandler) Target() event.Type {
	return event.ExecuteEventType
}

func (h *executeHandler) Handle(e event.Event, push func(event.Event)) error {
	executeEvent, isType := e.(*event.ExecuteEvent)
	if !isType {
		return event.ErrInvalidEventType
	}

	output, err := h.executor.Execute(executeEvent.Command)
	if err != nil {
		// NOTE: maybe some error should be handled by returning error spec data
		return fmt.Errorf("execute failed: %w", err)
	}

	push(&event.FormatEvent{
		ID_:  executeEvent.ID(),
		Data: output,
	})
	return nil
}
