package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"

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

	switch cmdStr {
	case "PING":
		return &spec.PingCommand{}, nil

	case "ECHO":
		echoCmd, err := p.parseEchoCommand(data)
		if err != nil {
			return nil, fmt.Errorf("invalid format for ECHO command: %w", err)
		}

		return echoCmd, nil

	case "GET":
		getCmd, err := p.parseGetCommand(data)
		if err != nil {
			return nil, fmt.Errorf("invalid format for GET command: %w", err)
		}

		return getCmd, nil

	case "SET":
		setCmd, err := p.parseSetCommand(data)
		if err != nil {
			return nil, fmt.Errorf("invalid format for SET command: %w", err)
		}

		return setCmd, nil

	default:
		return nil, fmt.Errorf("invalid command: %s", cmdStr)
	}
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

func (p *Parser) parseEchoCommand(data spec.Data) (*spec.EchoCommand, error) {
	arrData, isType := data.(*spec.ArrayData)
	if !isType {
		return nil, fmt.Errorf("data type %T is not array data type", data)
	}

	if arrData.Len != 2 {
		return nil, fmt.Errorf(
			"invalid ECHO command format: expected 1 argument, got %d",
			arrData.Len-1,
		)
	}

	var value string
	echoData := arrData.A[1]
	echoVal := echoData.Value()
	switch echoData.Type().Kind() {
	case reflect.String:
		value = echoVal.(string)
	case reflect.Int64:
		value = strconv.FormatInt(echoVal.(int64), 10)
	default:
		return nil, fmt.Errorf("invalid data type for echo command: %T", echoVal)
	}

	return &spec.EchoCommand{Value: value}, nil
}

func (p *Parser) parseGetCommand(data spec.Data) (*spec.GetCommand, error) {
	arrData, isType := data.(*spec.ArrayData)
	if !isType {
		return nil, fmt.Errorf("data type %T is not array data type", data)
	}

	if arrData.Len != 2 {
		return nil, fmt.Errorf(
			"invalid GET command format: expected 1 argument, got %d",
			arrData.Len-1,
		)
	}

	keyData := arrData.A[1]
	key, err := spec.Value[string](keyData)
	if err != nil {
		return nil, fmt.Errorf("invalid key type for GET command: %w", err)
	}

	return &spec.GetCommand{Key: key}, nil
}

func (p *Parser) parseSetCommand(data spec.Data) (*spec.SetCommand, error) {
	arrData, isType := data.(*spec.ArrayData)
	if !isType {
		return nil, fmt.Errorf("data type %T is not array data type", data)
	}

	// NOTE: after supporting extra options, this should be changed
	if arrData.Len < 2 {
		return nil, fmt.Errorf(
			"invalid SET command format: expected at least 2 arguments, got %d",
			arrData.Len-1,
		)
	}

	setCmd := &spec.SetCommand{}

	// parse key, value
	keyData := arrData.A[1]
	valueData := arrData.A[2]

	key, err := spec.Value[string](keyData)
	if err != nil {
		return nil, fmt.Errorf("invalid key type for SET command: %w", err)
	}
	setCmd.Key = key

	value, err := spec.Value[string](valueData)
	if err != nil {
		return nil, fmt.Errorf("invalid value type for SET command: %w", err)
	}
	setCmd.Value = value

	// parse flags
	const (
		FlagUninitialized = "\r\n" // initialize state, this is not a real flag
		FlagPX            = "PX"
	)
	flagState := FlagUninitialized
	for _, data := range arrData.A[3:] {
		switch flagState {
		case FlagUninitialized:
			flag, err := spec.Value[string](data)
			if err != nil {
				return nil, fmt.Errorf("invalid flag type for SET command: %w", err)
			}

			flagState = strings.ToUpper(flag)
		case FlagPX:
			millStr, err := spec.Value[string](data)
			if err != nil {
				return nil, fmt.Errorf("invalid PX value type for SET command: %w", err)
			}

			mill, err := strconv.ParseInt(millStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid PX value for SET command: %w", err)
			}

			if setCmd.ExpireAt != nil {
				return nil, fmt.Errorf("Expiration flag is already set for key %s", setCmd.Key)
			}

			expireAt := time.Now().Add(time.Duration(mill) * time.Millisecond)
			setCmd.ExpireAt = &expireAt

			flagState = FlagUninitialized
		}
	}

	return setCmd, nil
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
