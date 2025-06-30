package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/spec"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseHandler() *parseHandler {
	return &parseHandler{
		parser: p,
	}
}

func (p *Parser) Parse(data spec.Data) (spec.Command, error) {
	cmdStr, err := p.parseCommandString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	var cmd spec.Command
	switch cmdStr {
	case "PING":
		cmd = &spec.PingCommand{}
	case "ECHO":
		echoCmd := &spec.EchoCommand{}
		if err := p.fillEchoCommand(echoCmd, data); err != nil {
			return nil, fmt.Errorf("invalid format for ECHO command: %w", err)
		}
		cmd = echoCmd
	default:
		return nil, fmt.Errorf("invalid command: %s", cmdStr)
	}

	return cmd, nil
}

func (p *Parser) parseCommandString(data spec.Data) (string, error) {
	switch data.(type) {
	case *spec.SimpleStringData, *spec.BulkStringData:
		cmdStr, _ := spec.Value[string](data)
		return strings.ToUpper(cmdStr), nil
	case *spec.ArrayData:
		arr, _ := spec.Value[[]spec.Data](data)
		if len(arr) == 0 {
			return "", errors.New("array data is empty")
		}

		return p.parseCommandString(arr[0])
	default:
		return "", fmt.Errorf("data type %T is not contains command string", data)
	}
}

func (p *Parser) fillEchoCommand(cmd *spec.EchoCommand, data spec.Data) error {
	arrData, isType := data.(*spec.ArrayData)
	if !isType {
		return fmt.Errorf("data type %T is not array data type", data)
	}

	if arrData.Len >= 2 {
		echoData := arrData.A[1]
		echoVal := echoData.Value()
		switch echoData.Type().Kind() {
		case reflect.String:
			cmd.Value = echoVal.(string)
		case reflect.Int64:
			cmd.Value = strconv.FormatInt(echoVal.(int64), 10)
		default:
			return fmt.Errorf("invalid data type for echo command: %T", echoVal)
		}
	}

	return nil
}

type parseHandler struct {
	parser *Parser
}

func (h *parseHandler) Target() event.Type {
	return event.ParseEventType
}

func (h *parseHandler) Handle(e event.Event, push func(event.Event)) error {
	parseEvent, isType := e.(*event.ParseEvent)
	if !isType {
		return event.ErrInvalidEventType
	}

	cmd, err := h.parser.Parse(parseEvent.Data)
	if err != nil {
		return fmt.Errorf("parse error for data: %+v, err: %w", parseEvent.Data, err)
	}
	slog.Info("parsed to...",
		slog.Uint64("id", parseEvent.ID()),
		slog.Any("command", cmd),
	)

	push(&event.ExecuteEvent{
		ID_:     parseEvent.ID(),
		Command: cmd,
	})

	return nil
}
