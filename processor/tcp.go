package processor

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/codecrafters-io/redis-starter-go/event"
)

type TCPProcessor struct {
	listener net.Listener
	logger   *slog.Logger

	pushStopSignal chan struct{}
}

func NewTCPProcessor(address string, logger *slog.Logger) (*TCPProcessor, error) {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return nil, fmt.Errorf("failed to bind to address %s", address)
	}

	return &TCPProcessor{
		listener:       l,
		logger:         logger,
		pushStopSignal: make(chan struct{}),
	}, nil
}

func (t *TCPProcessor) Close() error {
	if err := t.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}

	return nil
}

func (t *TCPProcessor) InitPushing(push func(event.Event)) {
	go func() {
		for {
			select {
			case <-t.pushStopSignal:
				return
			default:
				conn, err := t.listener.Accept()
				if err != nil {
					t.logger.Error("failed to accept connection", "error", err)
					continue
				}

				push(&event.ReadEvent{Reader: conn})
			}
		}
	}()
}

func (t *TCPProcessor) ShutdownPushing() {
	if t.pushStopSignal != nil {
		close(t.pushStopSignal)
	}
}

type TCPReadHandler struct{}

func (h *TCPReadHandler) Target() event.Type {
	return event.ReadEventType
}

func (h *TCPReadHandler) Handle(e event.Event, push func(event.Event)) error {
	readEvent, ok := e.(*event.ReadEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	conn, ok := readEvent.Reader.(net.Conn)
	if !ok {
		return fmt.Errorf("reader is not a net.Conn: %T", readEvent.Reader)
	}

	go func() {
		bufReader := bufio.NewReader(conn)
		if _, err := bufReader.ReadString('\n'); err != nil {
			if errors.Is(err, io.EOF) {
				push(&event.CloseEvent{Closer: conn})
			} else {
				push(&event.ErrorEvent{Err: fmt.Errorf("failed to read from reader: %w", err)})
			}
			return
		}

		// NOTE: should be changed to parse command
		push(&event.WriteEvent{
			Data:   []byte("+PONG\r\n"),
			Writer: conn,
		})
	}()

	return nil
}

type TCPWriteHandler struct{}

func (h *TCPWriteHandler) Target() event.Type {
	return event.WriteEventType
}

func (h *TCPWriteHandler) Handle(e event.Event, push func(event.Event)) error {
	writeEvent, ok := e.(*event.WriteEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	conn, ok := writeEvent.Writer.(net.Conn)
	if !ok {
		return fmt.Errorf("invalid connection type")
	}

	go func() {
		writer := bufio.NewWriter(conn)
		_, err := writer.Write(writeEvent.Data)
		if err != nil {
			push(&event.ErrorEvent{Err: err})
			return
		}

		err = writer.Flush()
		if err != nil {
			push(&event.ErrorEvent{Err: err})
			return
		}

		push(&event.ReadEvent{Reader: conn})
	}()

	return nil
}

type TCPCloseHandler struct{}

func (h *TCPCloseHandler) Target() event.Type {
	return event.CloseEventType
}

func (h *TCPCloseHandler) Handle(e event.Event, push func(event.Event)) error {
	closeEvent, ok := e.(*event.CloseEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	conn, ok := closeEvent.Closer.(net.Conn)
	if !ok {
		return fmt.Errorf("invalid connection type")
	}

	if err := conn.Close(); err != nil {
		push(&event.ErrorEvent{Err: fmt.Errorf("failed to close connection: %w", err)})
		return nil
	}

	return nil
}
