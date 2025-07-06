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

type GetCommand struct {
	Key string
}

func (e *GetCommand) command() {}

type SetCommand struct {
	Key   string
	Value string
}

func (e *SetCommand) command() {}
