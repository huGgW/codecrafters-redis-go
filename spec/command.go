package spec

type Command interface {
	command()
}

type PingCommand struct{}

func (e *PingCommand) command() {}

type EchoCommand struct {
	Value string
}

func (e *EchoCommand) command() {}
