package processor

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/spec"
	"github.com/codecrafters-io/redis-starter-go/storage"
)

type Executor struct {
	storage storage.Storage
}

func NewExecutor(storage storage.Storage) *Executor {
	return &Executor{
		storage: storage,
	}
}

func (e *Executor) ExecuteHandler() *executeHandler {
	return &executeHandler{
		executor: e,
	}
}

func (e *Executor) Execute(cmd spec.Command) (spec.Data, error) {
	switch cmd.(type) {
	case *spec.PingCommand:
		return spec.SimpleStringOf("PONG"), nil

	case *spec.EchoCommand:
		echoCmd := cmd.(*spec.EchoCommand)
		return spec.BulkStringOf(echoCmd.Value), nil

	case *spec.GetCommand:
		getCmd := cmd.(*spec.GetCommand)
		val, err := e.storage.Get(getCmd.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", getCmd.Key, err)
		}

		if val == nil {
			return spec.NullBulkString(), nil
		} else {
			return spec.BulkStringOf(*val), nil
		}

	case *spec.SetCommand:
		setCmd := cmd.(*spec.SetCommand)
		if err := e.storage.Set(setCmd.Key, setCmd.Value, setCmd.ExpireAt); err != nil {
			return nil, fmt.Errorf("failed to set key {%s} as value {%s}: %w", setCmd.Key, setCmd.Value, err)
		}

		return spec.SimpleStringOf("OK"), nil

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
