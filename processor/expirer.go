package processor

import (
	"log/slog"
	"time"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/id"
	"github.com/codecrafters-io/redis-starter-go/storage"
)

var _ event.Pusher = (*Expirer)(nil)

type Expirer struct {
	d        time.Duration
	idissuer id.IDIssuer[uint64]
	storage  storage.Storage

	t              *time.Ticker
	pushStopSignal chan struct{}
}

func NewExpirer(duration time.Duration, idIssuer id.IDIssuer[uint64], storage storage.Storage) *Expirer {
	return &Expirer{
		d:        duration,
		idissuer: idIssuer,
		storage:  storage,
	}
}

func (t *Expirer) InitPushing(push func(event.Event)) {
	t.pushStopSignal = make(chan struct{})
	t.t = time.NewTicker(t.d)
	go t.loop(push)
}

func (t *Expirer) ShutdownPushing() {
	if t.pushStopSignal != nil {
		close(t.pushStopSignal)
	}

	if t.t != nil {
		t.t.Stop()
	}
}

func (t *Expirer) loop(push func(event.Event)) {
	for {
		select {
		case time := <-t.t.C:
			id := t.idissuer.Issue()
			slog.Info("expirer ticker fired, pushing expire event",
				slog.Uint64("id", id),
				slog.Time("time", time),
			)
			push(&event.ExpireEvent{
				ID_:  id,
				Time: time,
			})
		case <-t.pushStopSignal:
			slog.Info("expirer shutdown signal received")
			return
		}
	}
}

func (t *Expirer) ExpireEventHandler() *expireEventHandler {
	return &expireEventHandler{
		e: t,
	}
}

func (t *Expirer) expire(time time.Time) {
	slog.Info("expiring expired entries",
		slog.Time("expiry_time", time),
	)
	t.storage.ExpireAllUntil(time)
}

type expireEventHandler struct {
	e *Expirer
}

func (h *expireEventHandler) Handle(e event.Event, push func(event.Event)) error {
	expireEvent, ok := e.(*event.ExpireEvent)
	if !ok {
		return event.ErrInvalidEventType
	}

	slog.Info("handling expire event",
		slog.Uint64("id", expireEvent.ID()),
		slog.Time("time", expireEvent.Time),
	)
	h.e.expire(expireEvent.Time)
	return nil
}

func (h *expireEventHandler) Target() event.Type {
	return event.ExpireEventType
}
