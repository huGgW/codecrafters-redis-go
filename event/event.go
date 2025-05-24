package event

type Type string

type Event interface {
	Type() Type
}

type Handler interface {
	Handle(event Event, push func(Event)) error
	Target() Type
}

type Pusher interface {
	InitPushing(push func(Event))
	ShutdownPushing()
}
