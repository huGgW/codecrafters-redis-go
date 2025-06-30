package processor

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/id"
	"github.com/codecrafters-io/redis-starter-go/pkg"
)

type TCPProcessor struct {
	listener net.Listener
	idIssuer id.IDIssuer[uint64]

	connMap        *pkg.ConcurrentMap[uint64, *connInfo]
	pushStopSignal chan struct{}
}

type connInfo struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

func NewTCPProcessor(address string, idIssuer id.IDIssuer[uint64]) (*TCPProcessor, error) {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		return nil, fmt.Errorf("failed to bind to address %s", address)
	}

	return &TCPProcessor{
		listener: l,
		idIssuer: idIssuer,

		connMap:        pkg.NewConcurrentMap[uint64, *connInfo](),
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
					slog.Error("failed to accept connection", "error", err)
					continue
				}

				id := t.idIssuer.Issue()

				t.connMap.Store(id, &connInfo{
					conn: conn,
				})

				push(&event.ReadEvent{ID_: id})
			}
		}
	}()
}

func (t *TCPProcessor) ShutdownPushing() {
	if t.pushStopSignal != nil {
		close(t.pushStopSignal)
	}
}

func (t *TCPProcessor) ReadHandler() *tcpReadHandler {
	return &tcpReadHandler{
		tcpProcessor: t,
	}
}

func (t *TCPProcessor) WriteHandler() *tcpWriteHandler {
	return &tcpWriteHandler{
		tcpProcessor: t,
	}
}

func (t *TCPProcessor) CloseHandler() *tcpCloseHandler {
	return &tcpCloseHandler{tcpProcessor: t}
}

type tcpReadHandler struct {
	tcpProcessor *TCPProcessor
}

func (h *tcpReadHandler) Target() event.Type {
	return event.ReadEventType
}

func (h *tcpReadHandler) Handle(e event.Event, push func(event.Event)) error {
	readEvent, ok := e.(*event.ReadEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	connInfo, ok := h.tcpProcessor.connMap.Load(readEvent.ID())
	if !ok {
		return fmt.Errorf("connection does not exists for id %d", readEvent.ID())
	}

	if connInfo.scanner == nil {
		connInfo.scanner = pkg.NewCRLFScanner(connInfo.conn)
	}

	go func() {
		scanned := connInfo.scanner.Scan()
		if !scanned {
			if err := connInfo.scanner.Err(); err != nil {
				push(&event.ErrorEvent{
					Event: readEvent,
					Err:   err,
				})
			}
			connInfo.scanner = nil
			push(&event.CloseEvent{ID_: readEvent.ID()})
			return
		}

		data := connInfo.scanner.Bytes()
		slog.Info("read from",
			slog.Uint64("id", readEvent.ID()),
			slog.Any("conn", connInfo.conn.RemoteAddr()),
			slog.Any("data", data),
		)
		push(&event.LexingEvent{
			ID_:  readEvent.ID(),
			Data: data,
		})
	}()

	return nil
}

type tcpWriteHandler struct {
	tcpProcessor *TCPProcessor
}

func (h *tcpWriteHandler) Target() event.Type {
	return event.WriteEventType
}

func (h *tcpWriteHandler) Handle(e event.Event, push func(event.Event)) error {
	writeEvent, ok := e.(*event.WriteEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	ci, ok := h.tcpProcessor.connMap.Load(writeEvent.ID())
	if !ok {
		return fmt.Errorf("connection does not exists for id %d", writeEvent.ID())
	}

	go func() {
		slog.Info("write to",
			slog.Uint64("id", writeEvent.ID()),
			slog.Any("conn", ci.conn.RemoteAddr()),
			slog.Any("data", writeEvent.Data),
		)

		writer := bufio.NewWriter(ci.conn)
		_, err := writer.Write(writeEvent.Data)
		if err != nil {
			push(&event.ErrorEvent{Event: writeEvent, Err: err})
			return
		}

		err = writer.Flush()
		if err != nil {
			push(&event.ErrorEvent{Event: writeEvent, Err: err})
			return
		}

		// maybe more data is available to read, so we always publish ReadEvent
		id := h.tcpProcessor.idIssuer.Issue()
		h.tcpProcessor.connMap.Store(
			id,
			&connInfo{
				conn:    ci.conn,
				scanner: ci.scanner,
			},
		)

		slog.Info("pushing read event after write",
			slog.Uint64("from_id", writeEvent.ID()),
			slog.Uint64("to_id", id),
			slog.Any("conn", ci.conn.RemoteAddr()),
		)
		push(&event.ReadEvent{ID_: id})
	}()

	return nil
}

type tcpCloseHandler struct {
	tcpProcessor *TCPProcessor
}

func (h *tcpCloseHandler) Target() event.Type {
	return event.CloseEventType
}

func (h *tcpCloseHandler) Handle(e event.Event, push func(event.Event)) error {
	closeEvent, ok := e.(*event.CloseEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	connInfo, loaded := h.tcpProcessor.connMap.LoadAndDelete(closeEvent.ID())
	if !loaded {
		return nil // when connection is already closed, do nothing
	}

	slog.Info("closing....",
		slog.Uint64("id", closeEvent.ID()),
		slog.Any("conn", connInfo.conn.RemoteAddr()),
	)
	if err := connInfo.conn.Close(); err != nil {
		push(&event.ErrorEvent{Event: closeEvent, Err: fmt.Errorf("failed to close connection: %w", err)})
		return nil
	}

	return nil
}
