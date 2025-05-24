package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/processor"
)

func main() {
	// add notifier
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	// initialize logger
	logger := slog.Default()

	logger.Info("Starting Redis server...")

	// initialize handlers
	tcpProcessor, err := processor.NewTCPProcessor("0.0.0.0:6379", logger)
	if err != nil {
		logger.Error("failed to initialicze tcp processor", "error", err)
		os.Exit(1)
	}
	defer func() { _ = tcpProcessor.Close() }()

	loop := event.NewLoop(
		[]event.Handler{
			&processor.TCPReadHandler{},
			&processor.TCPWriteHandler{},
			&processor.TCPCloseHandler{},
			processor.NewErrorHandler(logger),
		},
		[]event.Pusher{tcpProcessor},
		logger,
	)

	loop.Start()
	<-shutdownCh
	loop.Shutdown()
}

// func listen(listener net.Listener) error {
// 	for {
// 		if err := func() error {
// 			conn, err := listener.Accept()
// 			if err != nil {
// 				return fmt.Errorf("Error accepting connection: %w", err)
// 			}
// 			defer conn.Close()
//
// 			if err := handleConn(conn); err != nil {
// 				if errors.Is(err, io.EOF) {
// 					return nil
// 				}
// 				return fmt.Errorf("Error handling connection: %w", err)
// 			}
//
// 			return nil
// 		}(); err != nil {
// 			return err
// 		}
// 	}
// }
//
// func handleConn(conn net.Conn) error {
// 	reader := bufio.NewReader(conn)
// 	writer := bufio.NewWriter(conn)
//
// 	for {
// 		_, err := reader.ReadString('\n')
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				return io.EOF
// 			}
// 			return fmt.Errorf("Error read from connection: %w", err)
// 		}
//
// 		// TODO: add encoder for simple string
// 		_, err = writer.WriteString("+PONG\r\n")
// 		if err != nil {
// 			return fmt.Errorf("Error write to connection: %w", err)
// 		}
//
// 		if err := writer.Flush(); err != nil {
// 			return fmt.Errorf("Error write to connection: %w", err)
// 		}
// 	}
// }
