package event

import (
	"log/slog"
	"slices"
	"sync"
	"time"

	queue "github.com/codecrafters-io/redis-starter-go/pkg"
)

type Loop struct {
	handlerMap map[Type]Handler
	pushers    []Pusher

	queue      *queue.ConcurrentQueue[Event]
	stopSignal chan struct{}
	wg         *sync.WaitGroup
}

func NewLoop(
	handlers []Handler,
	pushers []Pusher,
) *Loop {
	handlerMap := make(map[Type]Handler, len(handlers))
	for _, h := range handlers {
		handlerMap[h.Target()] = h
	}

	return &Loop{
		handlerMap: handlerMap,
		pushers:    slices.Clone(pushers),
		queue:      queue.NewConcurrentQueue[Event](),
	}
}

func (l *Loop) Start() {
	l.stopSignal = make(chan struct{})
	l.wg = new(sync.WaitGroup)
	l.wg.Add(1)

	for _, p := range l.pushers {
		p.InitPushing(l.queue.Enqueue)
	}

	go l.loop()
}

func (l *Loop) Shutdown() {
	for _, p := range l.pushers {
		p.ShutdownPushing()
	}
	l.stopSignal <- struct{}{}
	l.wg.Wait()
}

func (l *Loop) loop() {
	defer l.wg.Done()
	for {
		select {
		case <-l.stopSignal:
			for {
				e, has := l.queue.Dequeue()
				if !has {
					return
				}
				l.handleEvent(e)
			}
		default:
			e, has := l.queue.Dequeue()
			if has {
				l.handleEvent(e)
			} else {
				time.Sleep(time.Microsecond)
			}
		}
	}
}

func (l *Loop) handleEvent(e Event) {
	h, ok := l.handlerMap[e.Type()]
	if !ok {
		slog.Error("no handler found for event", slog.Any("event", e))
		return
	}

	if err := h.Handle(e, l.queue.Enqueue); err != nil {
		slog.Error("failed to handle event", slog.Any("event", e), slog.Any("error", err))
		return
	}
}
