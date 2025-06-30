package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/spec"
)

type Lexer struct {
	stateMap map[uint64]*lexingState // lexing is only executed once at a time, so no need for concurrency
}

func NewLexer() *Lexer {
	return &Lexer{
		stateMap: make(map[uint64]*lexingState),
	}
}

func (p *Lexer) LexingHandler() *lexingHandler {
	return &lexingHandler{
		lexer: p,
	}
}

func (p *Lexer) Lexing(id uint64, data []byte) (*lexingState, error) {
	var err error

	lexingState, found := p.stateMap[id]
	if !found {
		lexingState, err = p.beginLexing(data)
		if err != nil {
			return nil, fmt.Errorf("failed to lexing: %w", err)
		}

		if !lexingState.complete() {
			p.stateMap[id] = lexingState
		}
	} else {
		lexingState, err = p.continueLexing(data, lexingState)
		if err != nil {
			delete(p.stateMap, id)
			return nil, fmt.Errorf("failed to lexing: %w", err)
		}

		if lexingState.complete() {
			delete(p.stateMap, id)
		}
	}

	return lexingState, nil
}

func (p *Lexer) beginLexing(data []byte) (*lexingState, error) {
	typeId := data[0]
	body := data[1:]

	redisData, err := p.lexingOf(typeId, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create data from type %c: %w", typeId, err)
	}

	if redisData.Incomplete() {
		return &lexingState{
			incompleteDataStack: []spec.Data{redisData},
		}, nil
	} else {
		return &lexingState{completeData: redisData}, nil
	}
}

func (p *Lexer) continueLexing(data []byte, state *lexingState) (*lexingState, error) {
	if state.complete() {
		return state, nil
	}

	// pop stack
	continueData := state.incompleteDataStack[len(state.incompleteDataStack)-1]
	state.incompleteDataStack = state.incompleteDataStack[:len(state.incompleteDataStack)-1]

	// handle data
	switch (continueData).(type) {
	case *spec.BulkStringData:
		bulkString := continueData.(*spec.BulkStringData)
		if bulkString.Len != len(data) {
			return nil, fmt.Errorf("bulk string length mismatch: expected %d, got %d", bulkString.Len, len(data))
		}

		bulkString.S = string(data)
		state.completeData = bulkString

	case *spec.ArrayData:
		array := continueData.(*spec.ArrayData)
		elem, err := p.lexingOf(data[0], data[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to create array element from data: %w", err)
		}
		array.A = append(array.A, elem)

		switch {
		case elem.Incomplete():
			state.incompleteDataStack = append(state.incompleteDataStack, array)
			state.incompleteDataStack = append(state.incompleteDataStack, elem)
		case len(array.A) < array.Len:
			state.incompleteDataStack = append(state.incompleteDataStack, array)
		}
	}

	// handle nested completion case (ex. array)
	for !state.complete() {
		lastData := state.incompleteDataStack[len(state.incompleteDataStack)-1]
		if lastData.Incomplete() {
			break
		}

		state.incompleteDataStack = state.incompleteDataStack[:len(state.incompleteDataStack)-1]
		state.completeData = lastData
	}

	return state, nil
}

func (p *Lexer) lexingOf(typeId byte, body []byte) (spec.Data, error) {
	switch typeId {
	case SimpleStringPrefix:
		return &spec.SimpleStringData{S: string(body)}, nil

	case SimpleErrorPrefix:
		return &spec.SimpleErrorData{Err: errors.New(string(body))}, nil

	case IntegerPrefix:
		i, err := strconv.ParseInt(string(body), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to read integer: %w", err)
		}

		return &spec.IntegerData{I: i}, nil

	case BulkStringPrefix:
		l, err := strconv.Atoi(string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string length: %w", err)
		}

		return &spec.BulkStringData{Len: l}, nil

	case ArrayPrefix:
		l, err := strconv.Atoi(string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to read arrays length: %w", err)
		}

		return &spec.ArrayData{Len: l, A: make([]spec.Data, 0, l)}, nil

	default:
		return nil, errors.New("invalid first byte: " + string(typeId))
	}
}

type lexingState struct {
	incompleteDataStack []spec.Data
	completeData        spec.Data
}

func (s *lexingState) complete() bool {
	return len(s.incompleteDataStack) == 0
}

type lexingHandler struct {
	lexer *Lexer
}

func (h *lexingHandler) Target() event.Type {
	return event.LexingEventType
}

func (h *lexingHandler) Handle(e event.Event, push func(event.Event)) error {
	lexingEvent, ok := e.(*event.LexingEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	state, err := h.lexer.Lexing(lexingEvent.ID(), lexingEvent.Data)
	if err != nil {
		push(&event.CloseEvent{ID_: lexingEvent.ID()})
		return fmt.Errorf("fail to handle lexing event: %w", err)
	}

	if state.complete() {
		slog.Info(
			"lexing complete",
			slog.Uint64("id", lexingEvent.ID()),
			slog.Any("completeData", state.completeData),
		)
		push(&event.ParseEvent{
			ID_:  lexingEvent.ID(),
			Data: state.completeData,
		})
	} else {
		slog.Info(
			"lexing in progress",
			slog.Uint64("id", lexingEvent.ID()),
			slog.Any("incompleteDataStackTop", state.incompleteDataStack[len(state.incompleteDataStack)-1]),
		)
		push(&event.ReadEvent{ID_: lexingEvent.ID()})
	}

	return nil
}
