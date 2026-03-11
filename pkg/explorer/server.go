package explorer

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/nxadm/tail"

	"github.com/crbednarz/df-explorer/pkg/docker"
	"github.com/crbednarz/df-explorer/pkg/util"
	"github.com/docker/docker/client"
)

type Server struct {
	localSessionDir  string
	remoteSessionDir string
}

type logEntryJSON struct {
	Command    string `json:"command"`
	Operation  string `json:"operation"`
	Status     string `json:"status"`
	ReturnCode int    `json:"rc,omitempty"`
}

type CommandCallback func(ServerEvent) error

type ServerEvent struct {
	Command    string
	Operation  OperationType
	State      CommandState
	ReturnCode int
}

//go:embed df-env.sh
var envScript string

func newServer() (*Server, error) {
	cacheDir, err := util.CacheDir()
	if err != nil {
		return nil, err
	}

	sessionDir, err := os.MkdirTemp(cacheDir, "session-*")
	if err != nil {
		return nil, fmt.Errorf("unable to create session directory: %w", err)
	}
	historyFile, err := os.Create(path.Join(sessionDir, "history.log"))
	if err != nil {
		return nil, fmt.Errorf("unable to create history log file: %w", err)
	}
	historyFile.Close()

	envScriptName := "df-env.sh"
	err = os.WriteFile(path.Join(sessionDir, envScriptName), []byte(envScript), 0o644)
	if err != nil {
		return nil, fmt.Errorf("unable to write profile script: %w", err)
	}

	server := &Server{
		localSessionDir:  sessionDir,
		remoteSessionDir: "/tmp/df-explorer",
	}
	return server, nil
}

func (s *Server) SpawnContainer(ctx context.Context, cli *client.Client, image string) (docker.Container, error) {
	container, err := docker.NewContainer(
		ctx,
		cli,
		image,
		docker.WithMount(
			s.localSessionDir,
			s.remoteSessionDir,
		),
		docker.WithCommand([]string{"/bin/bash"}),
		docker.WithAttach(true),
	)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(container.Attachment(), "source %s\nreset\n", path.Join(s.remoteSessionDir, "df-env.sh"))
	return container, err
}

func (s *Server) historyLogPath() string {
	return path.Join(s.localSessionDir, "history.log")
}

func (s *Server) Listen(ctx context.Context, callback CommandCallback) error {
	file, err := tail.TailFile(s.historyLogPath(), tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return fmt.Errorf("unable to tail log file: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, ok := <-file.Lines:
			if !ok {
				return nil
			}
			var logEntry logEntryJSON
			err := json.Unmarshal([]byte(line.Text), &logEntry)
			if err != nil {
				return fmt.Errorf("unable to parse log entry: %w", err)
			}
			callback(eventFromLogEntry(logEntry))
		}
	}
}

func (s *Server) Close() error {
	return os.Remove(s.localSessionDir)
}

func eventFromLogEntry(entry logEntryJSON) ServerEvent {
	var state CommandState
	switch entry.Status {
	case "running":
		state = CommandStateRunning
	case "complete":
		state = CommandStateSuccess
	case "error":
		state = CommandStateError
	}

	return ServerEvent{
		Command:    entry.Command,
		Operation:  OperationType(entry.Operation),
		State:      state,
		ReturnCode: entry.ReturnCode,
	}
}
