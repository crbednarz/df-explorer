package explorer

type OperationType string

const (
	OperationNone        OperationType = ""
	OperationAdd         OperationType = "ADD"
	OperationArg         OperationType = "ARG"
	OperationCmd         OperationType = "CMD"
	OperationCopy        OperationType = "COPY"
	OperationEntrypoint  OperationType = "ENTRYPOINT"
	OperationEnv         OperationType = "ENV"
	OperationExpose      OperationType = "EXPOSE"
	OperationFrom        OperationType = "FROM"
	OperationHealthcheck OperationType = "HEALTHCHECK"
	OperationLabel       OperationType = "LABEL"
	OperationMaintainer  OperationType = "MAINTAINER"
	OperationOnbuild     OperationType = "ONBUILD"
	OperationRun         OperationType = "RUN"
	OperationShell       OperationType = "SHELL"
	OperationStopsignal  OperationType = "STOPSIGNAL"
	OperationUser        OperationType = "USER"
	OperationVolume      OperationType = "VOLUME"
	OperationWorkdir     OperationType = "WORKDIR"
)

type CommandState int

const (
	CommandStateSuccess CommandState = iota
	CommandStateError
	CommandStateRunning
)

type Command struct {
	Text  string
	State CommandState
}

func NewCommand(text string, state CommandState) Command {
	return Command{
		Text:  text,
		State: state,
	}
}
