package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/id"
	"github.com/codecrafters-io/redis-starter-go/processor"
	"github.com/codecrafters-io/redis-starter-go/storage"
)

func main() {
	// add notifier
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	// initialize logger
	slog.Info("Starting Redis server...")

	// initialize ID issuer
	idIssuer := &id.NumIDIssuer{}

	// initialize storage
	storage := storage.NewInMemoryStorage()

	// initialize handlers
	tcpProcessor, err := processor.NewTCPProcessor("0.0.0.0:6379", idIssuer)
	if err != nil {
		slog.Error("failed to initialicze tcp processor", "error", err)
		os.Exit(1)
	}
	defer func() { _ = tcpProcessor.Close() }()

	lexer := processor.NewLexer()
	parser := processor.NewParser()
	executor := processor.NewExecutor(storage)
	formatter := processor.NewFormatter()

	loop := event.NewLoop(
		[]event.Handler{
			tcpProcessor.ReadHandler(),
			tcpProcessor.WriteHandler(),
			tcpProcessor.CloseHandler(),
			lexer.LexingHandler(),
			parser.ParseHandler(),
			executor.ExecuteHandler(),
			formatter.FormatHandler(),
			processor.NewErrorHandler(),
		},
		[]event.Pusher{tcpProcessor},
	)

	loop.Start()
	<-shutdownCh
	loop.Shutdown()
}
